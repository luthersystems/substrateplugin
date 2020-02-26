package substratepluginshare

import (
	"context"
)

type transactionIDCtxKey struct{}

type transactionIDValue struct {
	txID string
}

// ContextWithTransactionID adds a unique value to a context for storing a
// transaction ID
func ContextWithTransactionID(ctx context.Context) context.Context {
	return context.WithValue(ctx, transactionIDCtxKey{}, &transactionIDValue{})
}

// SetContextTransactionID sets the transaction ID in a context value that has
// been initialized using ContextWithTransactionID
func SetContextTransactionID(ctx context.Context, txID string) {
	if val, ok := ctx.Value(transactionIDCtxKey{}).(*transactionIDValue); ok {
		val.txID = txID
	}
}

// GetContextTransactionID gets the transaction ID from a context value if present
func GetContextTransactionID(ctx context.Context) string {
	if val, ok := ctx.Value(transactionIDCtxKey{}).(*transactionIDValue); ok {
		return val.txID
	}
	return ""
}
