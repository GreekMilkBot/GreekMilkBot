package gmb

import (
	"context"
	"fmt"
	"github.com/GreekMilkBot/GreekMilkBot/adapter"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"sync"
)

type BotMessageHandle func(ctx context.Context, message bot.Message)
type BotEventHandle func(ctx context.Context, message bot.Event)

type GreekMilkBot struct {
	config *Config

	rx chan bot.Packet // 响应事件
	tx chan bot.Packet // 操作事件

	locker      *sync.RWMutex
	handleMsg   BotMessageHandle
	handleEvent BotEventHandle
}

func NewGreekMilkBot(config *Config) *GreekMilkBot {
	return &GreekMilkBot{
		config: config,
		locker: new(sync.RWMutex),

		tx: make(chan bot.Packet, config.Cache),
		rx: make(chan bot.Packet, config.Cache),
	}
}

func (g *GreekMilkBot) Run(ctx context.Context) error {
	bootCtx, cancel := context.WithCancel(ctx)
	gmap := make(map[string]adapter.Bus)
	for id, adapt := range g.config.Adapters {
		pid := fmt.Sprintf("%002d", id)
		bus := adapter.Bus{
			ID:      pid,
			Context: bootCtx,
			Tx:      g.rx,
			Rx:      make(chan bot.Packet, g.config.Cache),
		}
		gmap[pid] = bus
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
l:
	for {
		select {
		case <-ctx.Done():
			log.Debug("exiting bot loop")
			break l
		case event := <-g.rx:
			stat := gmap[event.Plugin]
			go func() {
				switch event.Type {
				case bot.PacketMessage:
					ctx := context.WithValue(ctx, "plugin", stat.ID)
					g.handleMsg(ctx, event.Data.(bot.Message))
				}
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
