package core

import (
	"context"
	"net/url"
)

var plugins = make(map[string]PluginHandler)

func GetAdapter(key string) (PluginHandler, bool) {
	handler, ok := plugins[key]
	return handler, ok
}

type Plugin interface {
	Bind(ctx *PluginBus) error
}

type PluginHandler func(ctx context.Context, url url.URL) (Plugin, error)

func RegisterAdapter(name string, adapter PluginHandler) {
	if plugins[name] != nil {
		panic("duplicate adapter name: " + name)
	}
	plugins[name] = adapter
}
