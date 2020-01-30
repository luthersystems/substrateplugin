package private

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"bitbucket.org/luthersystems/substrateplugin/substratepluginshare"
	"bitbucket.org/luthersystems/substrateplugin/substratepluginuser"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

const (
	// ShiroEndpointDecode is used to decode private data.
	ShiroEndpointDecode = "private_decode"
	// ShiroEndpointEncode is used to encode private data.
	ShiroEndpointEncode = "private_encode"
	// ShiroEndpointPurge is used to purge private data from the blockchain for
	// a data subject.
	ShiroEndpointPurge = "private_purge"
	// ShiroEndpointExport is used to export a data subject's private data.
	ShiroEndpointExport = "private_export"
	// ShiroEndpointProfileToDSID is used to get a DSID given a profile.
	ShiroEndpointProfileToDSID = "private_get_dsid"
)

const (
	hkdfSeedSize = 32
)

// SeedGen generates random secret keys. This is a hook that can be overridden
// at run time.
var SeedGen = func() ([]byte, error) {
	key := make([]byte, hkdfSeedSize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Encryptor selects message transform encryption algorithms.
type Encryptor string

// EncryptorNone indicates that no encryption should be applied.
const EncryptorNone Encryptor = "none"

// EncryptorAES256 indicates that AES-256 encryption should be applied.
const EncryptorAES256 Encryptor = "AES-256"

// Compressor selects message transform compression algortihms.
type Compressor string

// CompressorNone indicates that no compression should be applied.
const CompressorNone Compressor = "none"

// CompressorZlib indicates that zlib compression should be applied.
const CompressorZlib Compressor = "zlib"

// DSID is an identifier that represents a Data Subject.
type DSID string

// TransformHeader is a header for a message transformation.
// This is exported for json serialization.
type TransformHeader struct {
	// ProfilePaths are elpspaths that compose a data subject profile.
	ProfilePaths []string `json:"profile_paths"`
	// PrivatePaths are elpspaths that select private data.
	PrivatePaths []string `json:"private_paths"`
	// Encryptor selects the encryption algorithm.
	Encryptor Encryptor `json:"encryptor"`
	// Compressor selects the compression algorithm.
	Compressor Compressor `json:"compressor"`
}

// TransformBody is the body portion of a transformation. This is populated
// on encoded messages.
// This is exported for json serialization.
type TransformBody struct {
	// DSID is the data subject ID for the encoded transformation.
	DSID DSID `json:"dsid"`
	// EncryptedBase64 is the encrypted bytes belonging to the data subject.
	EncryptedBase64 string `json:"encrypted_base64"`
}

// Transform is a message transformation. It encapsulates both transformed
// messages (body), as well as settings to perform a transformation (header).
type Transform struct {
	// ContextPath represents an elpspath within the message where the
	// transformation will be applied. All transformation paths are relative
	// to this context.
	ContextPath string `json:"context_path"`
	// Header represents a transformation header. It is a description of
	// the transformation used for encoding and decoding.
	Header *TransformHeader `json:"header"`
	// Body includes an encoded message, where the encoding used the settings
	// defined in the Header.
	Body *TransformBody `json:"body"`
}

// EncodedMessage is a message that has undergone encoding.
// This is exported for json serialization.
type EncodedMessage struct {
	// MXF is a sentinel to indicate the message was encoded using libmxf.
	MXF string `json:"mxf"`
	// Message is the plaintext part of an encoded message.
	Message interface{} `json:"message"`
	// Transforms are the applied transforms.
	Transforms []*Transform `json:"transforms"`
}

// EncodeRequest is a request to encode a message.
// This is exported for json serialization.
type EncodeRequest struct {
	// Message is the message to be encoded.
	Message interface{} `json:"message"`
	// Transforms are the transformations to apply.
	Transforms []*Transform `json:"transforms"`
}

// EncodedResponse is a result of encoding a message, and can subsequently
// be decoded.
type EncodedResponse struct {
	// EncodedMessage is only set to the encoded message if encode did perform
	// encoding.
	encodedMessage *EncodedMessage
	// RawMessage is only set to the raw response if encode did not actually
	// perform any encoding.
	rawMessage *json.RawMessage
}

// MarshalJSON implements json.Marshaler.
func (r *EncodedResponse) MarshalJSON() ([]byte, error) {
	if r.encodedMessage == nil && r.rawMessage == nil {
		return nil, fmt.Errorf("empty response")
	}
	if r.encodedMessage == nil {
		return json.Marshal(r.rawMessage)
	}
	return json.Marshal(r.encodedMessage)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *EncodedResponse) UnmarshalJSON(b []byte) error {
	encMsg := &EncodedMessage{}
	err := json.Unmarshal(b, encMsg)
	if err != nil || encMsg.MXF == "" {
		raw := &json.RawMessage{}
		err = json.Unmarshal(b, raw)
		if err != nil {
			return err
		}
		r.rawMessage = raw
	} else {
		r.encodedMessage = encMsg
	}
	return nil
}

// WithParam returns a substratepluginshare config that passes a single parameter
// as an argument to an endpoint.
func WithParam(arg interface{}) substratepluginshare.Config {
	return substratepluginshare.WithParams([]interface{}{arg})
}

// WithSeed returns a substratepluginshare config that includes a CSPRNG seed.
func WithSeed() (substratepluginshare.Config, error) {
	seed, err := SeedGen()
	if err != nil {
		return nil, err
	}
	return substratepluginshare.WithTransientData("csprng_seed_private", seed), nil
}

// WithTransientMXF adds transient data used by MXF to encode and encrypt data.
// This config is not compatible with `WithTransientIVs`.
func WithTransientMXF(req *EncodeRequest) ([]substratepluginshare.Config, error) {
	if req == nil {
		req = &EncodeRequest{}
	}
	var configs []substratepluginshare.Config
	seedConfig, err := WithSeed()
	if err != nil {
		return nil, err
	}
	configs = append(configs, seedConfig)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	configs = append(configs, substratepluginshare.WithTransientData("mxf", reqBytes))
	return configs, nil
}

// Encode encodes a sensitive "message" using "transforms".
// If there no transforms, then encode simply returns a thin wrapper
// over the encoded message bytes.
func Encode(ctx context.Context, shiroHandle *substratepluginshare.Handle, message interface{}, transforms []*Transform, configs ...substratepluginshare.Config) (*EncodedResponse, error) {
	if len(transforms) == 0 {
		// fast path, nothing to do.
		rawBytes, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}
		encResp := &EncodedResponse{}
		err = json.Unmarshal(rawBytes, encResp)
		if err != nil {
			return nil, err
		}
		return encResp, nil
	}
	transientConfigs, err := WithTransientMXF(&EncodeRequest{
		Message:    message,
		Transforms: transforms,
	})
	if err != nil {
		return nil, err
	}
	configs = append(configs, transientConfigs...)
	resp, err := substratepluginuser.PluginCall(shiroHandle, ctx, ShiroEndpointEncode, configs)
	if err != nil {
		return nil, err
	}

	if resp.HasError {
		return nil, fmt.Errorf(resp.ErrorMessage)
	}
	enc := &EncodedResponse{}
	err = resp.UnmarshalTo(enc)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// Decode decodes a message that was encoded with transforms. If there are
// no transforms, then decode unmarshals the raw message bytes into "decoded".
func Decode(ctx context.Context, shiroHandle *substratepluginshare.Handle, encoded *EncodedResponse, decoded interface{}, configs ...substratepluginshare.Config) error {
	if encoded == nil {
		return fmt.Errorf("nil encoded message")
	}
	if encoded.encodedMessage == nil {
		// fast path, nothing to do.
		if encoded.rawMessage == nil {
			return fmt.Errorf("missing raw message")
		}
		rawBytes, err := json.Marshal(encoded.rawMessage)
		if err != nil {
			return err
		}
		message, ok := decoded.(proto.Message)
		if ok {
			return jsonpb.Unmarshal(bytes.NewReader(rawBytes), message)
		}
		return json.Unmarshal(rawBytes, decoded)
	}
	configs = append(configs, WithParam(encoded.encodedMessage))
	resp, err := substratepluginuser.PluginCall(shiroHandle, ctx, ShiroEndpointDecode, configs)
	if err != nil {
		return err
	}
	if resp.HasError {
		return fmt.Errorf(resp.ErrorMessage)
	}
	err = resp.UnmarshalTo(decoded)
	if err != nil {
		return err
	}
	return nil
}

// Export exports all sensitive data on the blockchain pertaining to a data
// subject with data subject ID "dsid".
func Export(ctx context.Context, shiroHandle *substratepluginshare.Handle, dsid DSID, exported map[string]interface{}, configs ...substratepluginshare.Config) error {
	if dsid == "" {
		return fmt.Errorf("invalid empty DSID")
	}
	configs = append(configs, WithParam(dsid))
	resp, err := substratepluginuser.PluginCall(shiroHandle, ctx, ShiroEndpointExport, configs)
	if err != nil {
		return err
	}
	if resp.HasError {
		return fmt.Errorf(resp.ErrorMessage)
	}
	err = resp.UnmarshalTo(exported)
	if err != nil {
		return err
	}
	return nil
}

// Purge removes all sensitive data on the blockchain pertaining to a data
// subject with data subject ID "dsid".
func Purge(ctx context.Context, shiroHandle *substratepluginshare.Handle, dsid DSID, configs ...substratepluginshare.Config) error {
	if dsid == "" {
		return fmt.Errorf("invalid empty DSID")
	}
	configs = append(configs, WithParam(dsid))
	seedConfig, err := WithSeed()
	if err != nil {
		return err
	}
	configs = append(configs, seedConfig)
	resp, err := substratepluginuser.PluginCall(shiroHandle, ctx, ShiroEndpointPurge, configs)
	if err != nil {
		return err
	}
	if resp.HasError {
		return fmt.Errorf(resp.ErrorMessage)
	}
	var gotDSID DSID
	err = resp.UnmarshalTo(gotDSID)
	if err != nil {
		return err
	}
	if gotDSID != dsid {
		return fmt.Errorf("unexpected response from purge: got %s != expected %s", gotDSID, dsid)
	}
	return nil
}

// ProfileToDSID returns a DSID for a data subject profile.
func ProfileToDSID(ctx context.Context, shiroHandle *substratepluginshare.Handle, profile interface{}, configs ...substratepluginshare.Config) (DSID, error) {
	configs = append(configs, WithParam(profile))
	resp, err := substratepluginuser.PluginCall(shiroHandle, ctx, ShiroEndpointProfileToDSID, configs)
	if err != nil {
		return "", err
	}
	if resp.HasError {
		return "", fmt.Errorf(resp.ErrorMessage)
	}
	var gotDSID DSID
	err = resp.UnmarshalTo(gotDSID)
	if err != nil {
		return "", err
	}
	return gotDSID, nil
}

// WrapCall wraps a shiro call. If the transaction logic encrypts new data
// then IVs must be specified, via the `WithTransientIVs` function.
// The configs passed to this are passed to the wrapped call, and not the
// encode and decode operations. This is to prevent the caller from accidently
// overwriting the transient data fields.
// IMPORTANT: The wrapper assumes the wrapped endpoint only takes a single
// argument!
func WrapCall(ctx context.Context, shiroHandle *substratepluginshare.Handle, method string, encTransforms ...*Transform) func(message interface{}, output interface{}, configs ...substratepluginshare.Config) error {
	return func(message interface{}, output interface{}, configs ...substratepluginshare.Config) error {
		encReq, err := Encode(ctx, shiroHandle, message, encTransforms, configs...)
		if err != nil {
			return fmt.Errorf("wrap encode error: %s", err)
		}
		configs = append([]substratepluginshare.Config{WithParam(encReq)}, configs...)
		resp, err := substratepluginuser.PluginCall(shiroHandle, ctx, method, configs)
		if err != nil {
			return fmt.Errorf("wrap call error: %s", err)
		}
		if resp.HasError {
			return fmt.Errorf("wrap call response error: %s", resp.ErrorMessage)
		}
		encResp := &EncodedResponse{}
		err = resp.UnmarshalTo(encResp)
		if err != nil {
			return err
		}
		err = Decode(ctx, shiroHandle, encResp, output, configs...)
		if err != nil {
			return fmt.Errorf("wrap decode error: %s", err)
		}
		return nil
	}
}
