package gmb

import (
	"context"

	"github.com/GreekMilkBot/GreekMilkBot/gmb/message"
)

type GreekMilkBot struct {
	config *Config

	handler *message.Handler
}

func NewGreekMilkBot(config *Config) *GreekMilkBot {
	return &GreekMilkBot{
		config:  config,
		handler: message.NewHandler(),
	}
}

func (g *GreekMilkBot) Run(ctx context.Context) error {
	for _, adapter := range g.config.Adapters {
		if err := adapter.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
