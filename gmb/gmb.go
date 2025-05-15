package gmb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type (
	BotMessageHandle func(ctx context.Context, message bot.Message)
	BotEventHandle   func(ctx context.Context, message bot.Event)
)

type GreekMilkBot struct {
	config *Config

	rx chan bot.Packet // 响应事件
	tx chan bot.Packet // 操作事件

	call *sync.Map // 事件回调

	handleMsg   BotMessageHandle
	handleEvent BotEventHandle
}

func NewGreekMilkBot(config *Config) *GreekMilkBot {
	return &GreekMilkBot{
		config: config,
		call:   new(sync.Map),

		tx: make(chan bot.Packet, config.Cache),
		rx: make(chan bot.Packet, config.Cache),
	}
}

func (g *GreekMilkBot) Run(ctx context.Context) error {
	bootCtx, cancel := context.WithCancel(ctx)
	gmap := make(map[string]*bot.Bus)
	for id, adapt := range g.config.Adapters {
		pid := fmt.Sprintf("%d", id)
		bus := bot.NewBus(pid, bootCtx, g.rx)
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

func (g *GreekMilkBot) loop(ctx context.Context, gmap map[string]*bot.Bus) {
l:
	for {
		select {
		case <-ctx.Done():
			log.Debug("exiting bot loop")
			break l
		case event := <-g.rx:
			stat := gmap[event.Plugin]
			ctx := context.WithValue(ctx, "gbot-plugin-id", stat.ID)
			go func() {
				switch event.Type {
				case bot.PacketMessage:
					go g.handleMsg(ctx, event.Data.(bot.Message))
				case bot.PacketAction:
					switch event.Data.(type) {
					case bot.Event:
						go g.handleEvent(ctx, event.Data.(bot.Event))
					case bot.ActionResponse:
						resp := event.Data.(bot.ActionResponse)
						if value, loaded := g.call.LoadAndDelete(resp.ID); loaded {
							vChan := value.(chan bot.ActionResponse)
							select {
							case vChan <- resp:
							case <-time.After(10 * time.Millisecond):
								log.Warn("timed out waiting for action response %s to channel", resp.ID)
							}
							close(vChan)
						} else {
							log.Warn("timed out waiting for action response %s", resp.ID)
						}
					}
				}
			}()
		case event := <-g.tx:
			stat := gmap[event.Plugin]
			switch event.Type {
			case bot.PacketAction:
				stat.Rx <- event.Data.(bot.ActionRequest)
			default:
				log.Error("unknown event type %s (or not support this type)", event.Type)
			}
		}
	}
}

func (g *GreekMilkBot) HandleMessageFunc(f BotMessageHandle) {
	g.handleMsg = f
}

func (g *GreekMilkBot) HandleEventFunc(f BotEventHandle) {
	g.handleEvent = f
}

func (g *GreekMilkBot) WithBot(ctx context.Context) *ClientBus {
	if id, ok := ctx.Value("gbot-plugin-id").(string); ok {
		return g.WithBotID(id)
	}
	return nil
}

func (g *GreekMilkBot) WithBotID(id string) *ClientBus {
	return &ClientBus{
		pluginID: id,
		tx:       g.tx,
		call:     g.call,
	}
}
