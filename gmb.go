package GreekMilkBot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"

	gmbcore "github.com/GreekMilkBot/GreekMilkBot/pkg"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/log"
)

type (
	BotMessageHandle func(ctx BotContext, message models.Message)
	BotEventHandle   func(ctx BotContext, event models.Event)
)

type WithPlugin struct {
	gmbcore.Plugin
	Bus       *gmbcore.PluginBus // 每个消息通信上下文
	Meta      *sync.Map          // 元数据
	Tools     mapset.Set[string] // 可用的工具包检查
	Resources *sync.Map          // 注册的资源解析器
}

func NewPlugin(p gmbcore.Plugin) *WithPlugin {
	return &WithPlugin{
		Plugin:    p,
		Meta:      &sync.Map{},
		Tools:     mapset.NewThreadUnsafeSet[string](),
		Resources: new(sync.Map),
	}
}

type GreekMilkBot struct {
	plugins []*WithPlugin
	cache   int

	rx chan models.Packet
	tx chan models.Packet

	call      *sync.Map
	handleMsg BotMessageHandle

	handleEvent BotEventHandle

	started *atomic.Bool
}

func NewGreekMilkBot(calls ...GMBConfig) (*GreekMilkBot, error) {
	init := &GreekMilkBot{
		call:    new(sync.Map),
		started: new(atomic.Bool),
		plugins: make([]*WithPlugin, 0),
		cache:   100,
	}
	for _, call := range calls {
		if err := call(init); err != nil {
			return nil, err
		}
	}
	init.tx = make(chan models.Packet, init.cache)
	init.rx = make(chan models.Packet, init.cache)

	return init, nil
}

func (g *GreekMilkBot) Run(ctx context.Context) error {
	if g.started.Load() {
		return errors.New("already started")
	}
	g.started.Store(true)
	bootCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	for gid, adapt := range g.plugins {
		adapt.Bus = gmbcore.NewAdapterBus(bootCtx, gid, g, func(packetType models.PacketType, data any) {
			g.rx <- models.Packet{
				Plugin: gid,
				Type:   packetType,
				Data:   data,
			}
		})
		if err := adapt.Bind(adapt.Bus); err != nil {
			return errors.Join(err, fmt.Errorf("plugin #%d Errorf", gid))
		}
	}
	return g.loop(ctx)
}

func (g *GreekMilkBot) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-g.rx:
			plugin := g.plugins[event.Plugin]
			go func() {
				ctx := BotContext{
					Context:   ctx,
					BotID:     event.Plugin,
					GMBSender: g,
					Tools:     plugin.Tools.ToSlice(),
					ResourceProviderFinderImpl: models.ResourceProviderFinderImpl{
						ResourceProviderManager: g,
					},
				}
				switch event.Type {
				case models.PacketMeta:
					meta := event.Data.(models.Meta)
					switch meta.Key {
					case ".tools":
						toolID := meta.Value
						if strings.Contains(toolID, ",") || strings.Contains(toolID, " ") {
							panic("tools id is not valid :" + toolID)
							return
						}
						if !plugin.Tools.Contains(meta.Value) {
							plugin.Tools.Add(meta.Value)
							log.Debugf("#%d add tool: %s , available tools: %v", event.Plugin, toolID, plugin.Tools.String())
						} else {
							log.Warnf("#%d tool[%s] already exists", event.Plugin, toolID)
						}
					default:
						if meta.Value != "" {
							plugin.Meta.Store(meta.Key, meta.Value)
						} else {
							plugin.Meta.Delete(meta.Key)
						}
					}
				case models.PacketMessage:
					go g.handleMsg(ctx, event.Data.(models.Message))
				case models.PacketAction:
					switch event.Data.(type) {
					case models.Event:
						go g.handleEvent(ctx, event.Data.(models.Event))
					case gmbcore.ActionResponse:
						resp := event.Data.(gmbcore.ActionResponse)
						if value, loaded := g.call.LoadAndDelete(resp.ID); loaded {
							vChan := value.(chan gmbcore.ActionResponse)
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
			plugin := g.plugins[event.Plugin]
			switch event.Type {
			case models.PacketAction:
				plugin.Bus.NewRequest(event.Data.(gmbcore.ActionRequest))
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

func (g *GreekMilkBot) Call(pluginID int, key string, params []any, result []any, timeout time.Duration) error {
	id := fmt.Sprintf("%d-%s-%d-%.2f", pluginID, key, time.Now().Unix(), rand.Float64())
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
	resultChan := make(chan gmbcore.ActionResponse)
	g.call.Store(id, resultChan)
	g.tx <- models.Packet{
		Plugin: pluginID,
		Type:   models.PacketAction,
		Data: gmbcore.ActionRequest{
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
			close(value.(chan gmbcore.ActionResponse))
			return errors.ErrUnsupported
		} else {
			return context.DeadlineExceeded
		}
	}
	return nil
}

func (g *GreekMilkBot) GetMeta(botID int, key string) (string, bool) {
	if load, ok := g.plugins[botID].Meta.Load(key); ok {
		return load.(string), ok
	}
	return "", false
}

func (g *GreekMilkBot) QueryResource(resource *models.Resource) (models.ResourceProvider, error) {
	plugin := g.plugins[resource.PluginID]
	if plugin == nil {
		return nil, errors.New(fmt.Sprintf("plugin not found: %d", resource.PluginID))
	}
	value, ok := plugin.Resources.Load(resource.Scheme)
	if !ok {
		return nil, errors.New("scheme not found:" + resource.Scheme)
	}
	return value.(models.ResourceProvider), nil
}

func (g *GreekMilkBot) RegisterResource(id int, scheme string, provider models.ResourceProvider) {
	plugin := g.plugins[id]
	if plugin == nil {
		panic(fmt.Sprintf("plugin id %d not found", id))
	}
	if _, loaded := plugin.Resources.LoadOrStore(scheme, provider); loaded {
		panic("scheme already registered: " + scheme)
	}
	log.Debugf("add resource scheme: %d#%s", id, scheme)
}

type GMBSender interface {
	Call(pluginID int, key string, params []any, result []any, timeout time.Duration) error
}
