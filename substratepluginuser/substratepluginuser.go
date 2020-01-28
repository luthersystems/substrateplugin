package substratepluginuser

import (
	"context"
	"fmt"
	"plugin"
	"runtime"

	"bitbucket.org/luthersystems/substrateplugin/substratepluginshare"
)

// Version is the plugin version to load.
var Version = "2.89.0-SNAPSHOT"

// PluginNewRPC is a pointer to the plugin's NewRPC function.
var PluginNewRPC func([]substratepluginshare.Config) *substratepluginshare.Handle

// PluginNewMock is a pointer to the plugin's NewRPC function.
var PluginNewMock func(string, string, string, []substratepluginshare.Config) (*substratepluginshare.Handle, error)

// PluginCall is a pointer to the plugin's NewRPC function.
var PluginCall func(*substratepluginshare.Handle, context.Context, string, []substratepluginshare.Config) (*substratepluginshare.Response, error)

// PluginIsTimeoutError is a pointer to the plugin's NewRPC function.
var PluginIsTimeoutError func(error) bool

// PluginQueryInfo is a pointer to the plugin's NewRPC function.
var PluginQueryInfo func(*substratepluginshare.Handle, []substratepluginshare.Config) (uint64, error)

// PluginQueryBlock is a pointer to the plugin's NewRPC function.
var PluginQueryBlock func(*substratepluginshare.Handle, uint64, []substratepluginshare.Config) (substratepluginshare.Block, error)

func init() {
	plugin, err := plugin.Open(fmt.Sprintf("substrateplugin-%s-%s-%s.so", runtime.GOOS, runtime.GOARCH, Version))
	if err != nil {
		panic(err)
	}

	rawNewRPC, err := plugin.Lookup("NewRPC")
	if err != nil {
		panic(err)
	}

	{
		var ok bool
		PluginNewRPC, ok = rawNewRPC.(func([]substratepluginshare.Config) *substratepluginshare.Handle)
		if !ok {
			panic("NewRPC")
		}
	}

	rawNewMock, err := plugin.Lookup("NewMock")
	if err != nil {
		panic(err)
	}

	{
		var ok bool
		PluginNewMock, ok = rawNewMock.(func(string, string, string, []substratepluginshare.Config) (*substratepluginshare.Handle, error))
		if !ok {
			panic("NewMock")
		}
	}

	rawCall, err := plugin.Lookup("Call")
	if err != nil {
		panic(err)
	}

	{
		var ok bool
		PluginCall, ok = rawCall.(func(*substratepluginshare.Handle, context.Context, string, []substratepluginshare.Config) (*substratepluginshare.Response, error))
		if !ok {
			panic("Call")
		}
	}

	rawIsTimeoutError, err := plugin.Lookup("IsTimeoutError")
	if err != nil {
		panic(err)
	}

	{
		var ok bool
		PluginIsTimeoutError, ok = rawIsTimeoutError.(func(error) bool)
		if !ok {
			panic("IsTimeoutError")
		}
	}

	rawQueryInfo, err := plugin.Lookup("QueryInfo")
	if err != nil {
		panic(err)
	}

	{
		var ok bool
		PluginQueryInfo, ok = rawQueryInfo.(func(*substratepluginshare.Handle, []substratepluginshare.Config) (uint64, error))
		if !ok {
			panic("QueryInfo")
		}
	}

	rawQueryBlock, err := plugin.Lookup("QueryBlock")
	if err != nil {
		panic(err)
	}

	{
		var ok bool
		PluginQueryBlock, ok = rawQueryBlock.(func(*substratepluginshare.Handle, uint64, []substratepluginshare.Config) (substratepluginshare.Block, error))
		if !ok {
			panic("QueryBlock")
		}
	}
}
