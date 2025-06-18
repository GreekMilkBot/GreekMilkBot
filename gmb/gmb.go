package gmb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type (
	BotMessageHandle func(ctx context.Context, id string, message bot.Message)
	BotEventHandle   func(ctx context.Context, id string, event bot.Event)

	ClientCallHandle func(pluginID string, key string, params []any, result []any, timeout time.Duration) error
)

type GreekMilkBot struct {
	config *Config

	rx chan bot.Packet
	tx chan bot.Packet

	call *sync.Map

	meta *sync.Map

	handleMsg   BotMessageHandle
	handleEvent BotEventHandle
}

func NewGreekMilkBot(calls ...GreekMilkBotConfig) (*GreekMilkBot, error) {
	config := DefaultConfig()
	for _, call := range calls {
		if err := call(config); err != nil {
			return nil, err
		}
	}
	init := &GreekMilkBot{
		config: config,
		call:   new(sync.Map),
		meta:   new(sync.Map),
		tx:     make(chan bot.Packet, config.Cache),
		rx:     make(chan bot.Packet, config.Cache),
	}
	return init, nil
}

func (g *GreekMilkBot) Run(ctx context.Context) error {
	bootCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	adapters := make(map[string]*bot.Bus)
	for gid, adapt := range g.config.Adapters {
		bus := bot.NewBus(fmt.Sprintf("%d", gid), bootCtx, g.rx)
		adapters[bus.ID] = bus
		if err := adapt.Bind(bus); err != nil {
			return errors.Join(err, fmt.Errorf("plugin #%d Errorf", gid))
		}
	}
	return g.loop(ctx, adapters)
}

func (g *GreekMilkBot) loop(ctx context.Context, gmap map[string]*bot.Bus) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-g.rx:
			stat := gmap[event.Plugin]
			go func() {
				switch event.Type {
				case bot.PacketMeta:
					meta := event.Data.(bot.Meta)
					metaGroup := &sync.Map{}
					if m, ok := g.meta.LoadOrStore(event.Plugin, metaGroup); !ok {
						metaGroup = m.(*sync.Map)
					}
					if meta.Value != "" {
						metaGroup.Store(meta.Key, meta.Value)
					} else {
						metaGroup.Delete(meta.Key)
					}
				case bot.PacketMessage:
					go g.handleMsg(ctx, stat.ID, event.Data.(bot.Message))
				case bot.PacketAction:
					switch event.Data.(type) {
					case bot.Event:
						go g.handleEvent(ctx, stat.ID, event.Data.(bot.Event))
					case bot.ActionResponse:
						resp := event.Data.(bot.ActionResponse)
						if value, loaded := g.call.LoadAndDelete(resp.ID); loaded {
							vChan := value.(chan bot.ActionResponse)
							select {
							case vChan <- resp:
							case <-time.After(10 * time.Millisecond):
								log.Warnf("timed out waiting for action response %s to channel", resp.ID)
							}
							close(vChan)
						} else {
							log.Warnf("timed out waiting for action response %s", resp.ID)
						}
					}
				}
			}()
		case event := <-g.tx:
			stat := gmap[event.Plugin]
			switch event.Type {
			case bot.PacketAction:
				stat.NewRequest(event.Data.(bot.ActionRequest))
			default:
				log.Errorf("unknown event type %s (or not support this type)", event.Type)
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

func (g *GreekMilkBot) ClientCall(pluginID string, key string, params []any, result []any, timeout time.Duration) error {
	id := fmt.Sprintf("%s-%s-%d-%.2f", pluginID, key, time.Now().Unix(), rand.Float64())
	for _, item := range result {
		paramType := reflect.TypeOf(item)
		if paramType.Kind() != reflect.Ptr {
			return errors.New("result must be a pointer")
		}
	}
	paramsRaw := make([]string, len(params))
	for i, param := range params {
		f, err := json.Marshal(param)
		if err != nil {
			return err
		}
		paramsRaw[i] = string(f)
	}
	resultChan := make(chan bot.ActionResponse)
	g.call.Store(id, resultChan)
	g.tx <- bot.Packet{
		Plugin: pluginID,
		Type:   bot.PacketAction,
		Data: bot.ActionRequest{
			ID:     id,
			Action: key,
			Params: paramsRaw,
		},
	}
	select {
	case res := <-resultChan:
		if res.OK {
			for i, item := range result {
				if err := json.Unmarshal([]byte(res.Data[i]), item); err != nil {
					return err
				}
			}
		} else {
			return errors.New(res.ErrorMsg)
		}
	case <-time.After(timeout):
		if value, loaded := g.call.LoadAndDelete(id); loaded {
			close(value.(chan bot.ActionResponse))
		}
		return context.DeadlineExceeded
	}
	return nil
}

func (g *GreekMilkBot) GetMeta(botID string) map[string]string {
	result := make(map[string]string)
	if value, ok := g.meta.Load(botID); ok {
		value.(*sync.Map).Range(func(k, v interface{}) bool {
			result[k.(string)] = v.(string)
			return true
		})
	}
	return result
}
