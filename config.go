package GreekMilkBot

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/GreekMilkBot/GreekMilkBot/pkg"
)

type GMBConfig func(*GreekMilkBot) error

func WithPlugins(plugins ...core.Plugin) GMBConfig {
	return func(config *GreekMilkBot) error {
		for _, adapter := range plugins {
			find := false
			for _, plugin := range config.plugins {
				if plugin.Plugin != adapter {
					find = true
					break
				}
			}
			if !find {
				config.plugins = append(config.plugins, NewPlugin(adapter))
			}
		}
		return nil
	}
}

func WithPluginURL(ctx context.Context, urlStr string) GMBConfig {
	return func(config *GreekMilkBot) error {
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
		return WithPlugins(b)(config)
	}
}
