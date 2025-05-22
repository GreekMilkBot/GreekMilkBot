package gmb

import (
	"context"
	"sync"

	"github.com/GreekMilkBot/GreekMilkBot/gmb/message"
)

type GMessage struct {
}

type GreekMilkBot struct {
	config *Config

	handler *message.Handler
	client  []chan GMessage
	locker  *sync.RWMutex
}

func NewGreekMilkBot(config *Config) *GreekMilkBot {
	return &GreekMilkBot{
		config:  config,
		handler: message.NewHandler(),
		locker:  new(sync.RWMutex),
		client:  make([]chan GMessage, 0),
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

func (g *GreekMilkBot) Receive() chan GMessage {
	g.locker.Lock()
	defer g.locker.Unlock()
	g.client = append(g.client, make(chan GMessage, g.config.Cache))
	panic("implement me")
}
