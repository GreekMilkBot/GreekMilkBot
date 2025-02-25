package greekmilkbot

import (
	"context"

	"github.com/GreekMilkBot/GreekMilkBot/adapter"
)

type Bot struct {
	Config   *Config
	Adapters []adapter.Adapter
}

func NewBot(config *Config) *Bot {
	return &Bot{
		Config:   config,
		Adapters: make([]adapter.Adapter, 0),
	}
}

func (b *Bot) AddAdapter(adapter adapter.Adapter) {
	b.Adapters = append(b.Adapters, adapter)
}

func (b *Bot) Run(ctx context.Context) error {
	for _, adapter := range b.Adapters {
		if err := adapter.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
