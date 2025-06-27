package GreekMilkBot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"

	core2 "github.com/GreekMilkBot/GreekMilkBot/pkg"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/log"
)

type (
	BotMessageHandle func(ctx BotContext, message models.Message)
	BotEventHandle   func(ctx BotContext, event models.Event)
)

type GreekMilkBot struct {
	config *Config

	rx chan models.Packet
	tx chan models.Packet

	call *sync.Map
	meta *sync.Map

	handleMsg   BotMessageHandle
	handleEvent BotEventHandle

	started *atomic.Bool
}

func NewGreekMilkBot(calls ...GMBConfig) (*GreekMilkBot, error) {
	config := DefaultConfig()
	for _, call := range calls {
		if err := call(config); err != nil {
			return nil, err
		}
	}
	init := &GreekMilkBot{
		config:  config,
		call:    new(sync.Map),
		meta:    new(sync.Map),
		tx:      make(chan models.Packet, config.Cache),
		rx:      make(chan models.Packet, config.Cache),
		started: new(atomic.Bool),
	}
	return init, nil
}

func (g *GreekMilkBot) Run(ctx context.Context) error {
	if g.started.Load() {
		return errors.New("already started")
	}
	g.started.Store(true)
	bootCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	adapters := make(map[string]*core2.AdapterBus)
	for gid, adapt := range g.config.Adapters {
		id := fmt.Sprintf("%d", gid)
		adapterBus := core2.NewAdapterBus(id, bootCtx, g.rx)
		g.meta.Store(adapterBus.ID, new(sync.Map))
		adapters[adapterBus.ID] = adapterBus
		if err := adapt.Bind(adapterBus); err != nil {
			return errors.Join(err, fmt.Errorf("plugin #%d Errorf", gid))
		}
	}
	return g.loop(ctx, adapters)
}

func (g *GreekMilkBot) loop(ctx context.Context, gmap map[string]*core2.AdapterBus) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-g.rx:
			stat := gmap[event.Plugin]
			go func() {
				tools, _ := g.GetMeta(event.Plugin, ".tools")
				ctx := BotContext{
					Context:   ctx,
					BotID:     stat.ID,
					GMBSender: g,
					Tools:     strings.Split(tools, ","),
				}
				switch event.Type {
				case models.PacketMeta:
					meta := event.Data.(models.Meta)
					var pluginMeta *sync.Map
					m, ok := g.meta.Load(event.Plugin)
					if !ok {
						log.Errorf("plugin %s not found", event.Plugin)
						return
					} else {
						pluginMeta = m.(*sync.Map)
					}
					switch meta.Key {
					case ".tools":
						g.addToolsMeta(event.Plugin, meta.Value, pluginMeta)
					default:
						if meta.Value != "" {
							pluginMeta.Store(meta.Key, meta.Value)
						} else {
							pluginMeta.Delete(meta.Key)
						}
					}
				case models.PacketMessage:
					go g.handleMsg(ctx, event.Data.(models.Message))
				case models.PacketAction:
					switch event.Data.(type) {
					case models.Event:
						go g.handleEvent(ctx, event.Data.(models.Event))
					case core2.ActionResponse:
						resp := event.Data.(core2.ActionResponse)
						if value, loaded := g.call.LoadAndDelete(resp.ID); loaded {
							vChan := value.(chan core2.ActionResponse)
							defer close(vChan)
							select {
							case vChan <- resp:
							case <-time.After(10 * time.Second):
								log.Warnf("timed out waiting for action response %s return", resp.ID)
							}
						} else {
							log.Warnf("timed out waiting for action response %s", resp.ID)
						}
					}
				}
			}()
		case event := <-g.tx:
			stat := gmap[event.Plugin]
			switch event.Type {
			case models.PacketAction:
				stat.NewRequest(event.Data.(core2.ActionRequest))
			default:
				log.Errorf("unknown event type %s (or not support this type)", event.Type)
			}
		}
	}
}

func (g *GreekMilkBot) addToolsMeta(pluginID, toolID string, pluginMeta *sync.Map) {
	if strings.Contains(toolID, ",") || strings.Contains(toolID, " ") {
		panic("tools id is not valid :" + toolID)
		return
	}
	for {
		tools := make([]string, 0)
		var old string
		if value, ok := pluginMeta.LoadOrStore(".tools", toolID); ok {
			old = value.(string)
			tools = append(tools, strings.Split(old, ",")...)
		} else {
			log.Debugf("#%s add tool: %s", pluginID, toolID)
			break
		}
		if slices.Contains(tools, toolID) {
			log.Warnf("#%s tool[%s] already exists", pluginID, toolID)
			break
		}
		tools = append(tools, toolID)
		if pluginMeta.CompareAndSwap(".tools", old, strings.Join(tools, ",")) {
			log.Debugf("#%s add tool: %s , available tools: %v", pluginID, toolID, tools)
			break
		}
	}
}

func (g *GreekMilkBot) HandleMessageFunc(f BotMessageHandle) {
	g.handleMsg = f
}

func (g *GreekMilkBot) HandleEventFunc(f BotEventHandle) {
	g.handleEvent = f
}

func (g *GreekMilkBot) Call(pluginID string, key string, params []any, result []any, timeout time.Duration) error {
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
	resultChan := make(chan core2.ActionResponse)
	g.call.Store(id, resultChan)
	g.tx <- models.Packet{
		Plugin: pluginID,
		Type:   models.PacketAction,
		Data: core2.ActionRequest{
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
			// 方法未被处理
			close(value.(chan core2.ActionResponse))
			return errors.ErrUnsupported
		} else {
			return context.DeadlineExceeded
		}
	}
	return nil
}

func (g *GreekMilkBot) GetMeta(botID string, key string) (string, bool) {
	if value, ok := g.meta.Load(botID); ok {
		if load, ok := value.(*sync.Map).Load(key); ok {
			return load.(string), ok
		}
	}
	return "", false
}

type GMBSender interface {
	Call(pluginID string, key string, params []any, result []any, timeout time.Duration) error
}
