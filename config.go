package GreekMilkBot

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/GreekMilkBot/GreekMilkBot/pkg"
)

type Config struct {
	Adapters []core.Adapter
	Cache    int
}

func DefaultConfig() *Config {
	return &Config{
		Adapters: make([]core.Adapter, 0),
		Cache:    100,
	}
}

type GMBConfig func(*Config) error

func WithAdapters(adapters ...core.Adapter) GMBConfig {
	return func(config *Config) error {
		for _, adapter := range adapters {
			if adapter != nil && !slices.Contains(config.Adapters, adapter) {
				config.Adapters = append(config.Adapters, adapter)
			}
		}
		return nil
	}
}

func WithAdapterURL(ctx context.Context, urlStr string) GMBConfig {
	return func(config *Config) error {
		sType, urlStr, found := strings.Cut(urlStr, "+")
		if !found {
			return fmt.Errorf("invalid url: %s", urlStr)
		}
		adapter, ok := core.GetAdapter(sType)
		if !ok {
			return fmt.Errorf("adapter not found: %s", sType)
		}
		u, err := url.Parse(urlStr)
		if err != nil {
			return err
		}
		b, err := adapter(ctx, *u)
		if err != nil {
			return err
		}
		return WithAdapters(b)(config)
	}
}
