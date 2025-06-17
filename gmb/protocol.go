package gmb

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
)

var adapters = make(map[string]AdapterHandler)

type AdapterHandler func(ctx context.Context, url url.URL) (bot.Adapter, error)

func RegisterAdapter(name string, adapter AdapterHandler) {
	if adapters[name] != nil {
		panic("duplicate adapter name: " + name)
	}
	adapters[name] = adapter
}

func WithAdapterURL(ctx context.Context, urlStr string) GreekMilkBotConfig {
	return func(config *Config) error {
		sType, urlStr, found := strings.Cut(urlStr, "+")
		if !found {
			return fmt.Errorf("invalid url: %s", urlStr)
		}
		adapter, ok := adapters[sType]
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
