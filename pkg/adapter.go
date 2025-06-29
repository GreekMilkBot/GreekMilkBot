package core

import (
	"context"
	"net/url"
)

var adapters = make(map[string]AdapterHandler)

func GetAdapter(key string) (AdapterHandler, bool) {
	handler, ok := adapters[key]
	return handler, ok
}

type Adapter interface {
	Bind(ctx *AdapterBus) error
}

type AdapterHandler func(ctx context.Context, url url.URL) (Adapter, error)

func RegisterAdapter(name string, adapter AdapterHandler) {
	if adapters[name] != nil {
		panic("duplicate adapter name: " + name)
	}
	adapters[name] = adapter
}
