package gmb

import (
	"slices"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
)

type Config struct {
	Adapters []bot.Adapter
	Cache    int
}

func DefaultConfig() *Config {
	return &Config{
		Adapters: make([]bot.Adapter, 0),
		Cache:    100,
	}
}

type GreekMilkBotConfig func(*Config) error

func WithAdapters(adapters ...bot.Adapter) GreekMilkBotConfig {
	return func(config *Config) error {
		for _, adapter := range adapters {
			if adapter != nil && !slices.Contains(config.Adapters, adapter) {
				config.Adapters = append(config.Adapters, adapter)
			}
		}
		return nil
	}
}
