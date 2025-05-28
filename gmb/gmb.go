package gmb

import (
	"context"
	"github.com/GreekMilkBot/GreekMilkBot/adapter"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"sync"
)

type BotMessageHandle func(ctx context.Context, message bot.Message)
type BotEventHandle func(ctx context.Context, message bot.Event)

type GreekMilkBot struct {
	config *Config

	rx chan Packet // 响应事件
	tx chan Packet // 操作事件

	locker      *sync.RWMutex
	handleMsg   BotMessageHandle
	handleEvent BotEventHandle
}

func NewGreekMilkBot(config *Config) *GreekMilkBot {
	return &GreekMilkBot{
		config: config,
		locker: new(sync.RWMutex),

		tx: make(chan Packet, config.Cache),
		rx: make(chan Packet, config.Cache),
	}
}

func (g *GreekMilkBot) Run(ctx context.Context) error {
	bootCtx, cancel := context.WithCancel(ctx)
	gmap := make(map[string]adapter.Bus)
	for _, adapt := range g.config.Adapters {
		id := adapt.ID()
		bus := adapter.Bus{
			ID:      id,
			Context: bootCtx,
		}
		gmap[id] = bus
		if err := adapt.Run(bus); err != nil {
			cancel()
			return err
		}
	}
	go func() {
		defer cancel()
		g.loop(ctx, gmap)
	}()
	return nil
}

func (g *GreekMilkBot) loop(ctx context.Context, gmap map[string]adapter.Bus) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("exiting bot loop")
			break
		case event := <-g.rx:
			adapt := gmap[event.plugin]
			go func() {

			}()
		case _ = <-g.tx:

		}
	}
}

func (g *GreekMilkBot) HandleMessageFunc(f BotMessageHandle) {
	g.locker.Lock()
	defer g.locker.Unlock()
	g.handleMsg = f
}

func (g *GreekMilkBot) HandleEventFunc(f BotEventHandle) {
	g.locker.Lock()
	defer g.locker.Unlock()
	g.handleEvent = f
}
