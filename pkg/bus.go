package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/log"
)

type ActionResponse struct {
	ID       string `json:"id"`
	OK       bool   `json:"ok"`
	ErrorMsg string `json:"error,omitempty"`

	Data []string `json:"data,omitempty"`
}

type ActionRequest struct {
	ID     string   `json:"id"`
	Action string   `json:"action"`
	Params []string `json:"params,omitempty"`
}

type PluginBus struct {
	context.Context

	ID       int // bot id
	send     func(packetType models.PacketType, data any)
	rx       chan ActionRequest
	call     *sync.Map
	resource models.ResourceProviderManager
	*models.ResourceProviderFinderImpl
}

func NewAdapterBus(ctx context.Context,
	id int,
	resources models.ResourceProviderManager,
	send func(packetType models.PacketType, data any),
) *PluginBus {
	bus := PluginBus{
		ID:       id,
		Context:  ctx,
		send:     send,
		rx:       make(chan ActionRequest, 100),
		call:     new(sync.Map),
		resource: resources,
		ResourceProviderFinderImpl: &models.ResourceProviderFinderImpl{
			Finder: resources,
		},
	}
	go bus.receiveLoop()
	return &bus
}

func (b *PluginBus) SendMessage(message models.Message) {
	b.send(models.PacketMessage, message)
}

func (b *PluginBus) SendEvent(event models.Event) {
	b.send(models.PacketAction, event)
}

func (b *PluginBus) SendMeta(key string, value string) {
	b.send(models.PacketMeta, models.Meta{
		Key:   key,
		Value: value,
	})
}

func (b *PluginBus) NewRequest(req ActionRequest) {
	b.rx <- req
}

func (b *PluginBus) receiveLoop() {
	defer b.call.Clear()
	defer close(b.rx)
	for {
		select {
		case <-b.Done():
			return
		case req := <-b.rx:
			if value, ok := b.call.Load(req.Action); ok {
				go b.exec(req, value)
			} else {
				b.sendError(req, errors.New("func not found"))
			}
		}
	}
}

func (b *PluginBus) exec(req ActionRequest, value any) {
	fnValue := reflect.ValueOf(value)
	fnType := fnValue.Type()
	if len(req.Params) != fnType.NumIn() {
		b.sendError(req, errors.New("invalid parameters count"))
		return
	}
	params := make([]reflect.Value, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		var paramValue reflect.Value
		if paramType.Kind() == reflect.Ptr {
			paramValue = reflect.New(paramType.Elem())
		} else {
			paramValue = reflect.New(paramType).Elem()
		}
		var err error
		if paramType.Kind() == reflect.Ptr {
			err = json.Unmarshal([]byte(req.Params[i]), paramValue.Interface())
		} else {
			err = json.Unmarshal([]byte(req.Params[i]), paramValue.Addr().Interface())
		}
		if err != nil {
			b.sendError(req, fmt.Errorf("unmarshal params[%d] err: %v", i, err))
			return
		}
		params[i] = paramValue
	}

	call := fnValue.Call(params)
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	results := make([]string, len(call))
	for i, v := range call {
		if v.Type().Implements(errorType) && !v.IsNil() {
			b.sendError(req, v.Interface().(error))
			return
		}
		marshal, err := json.Marshal(v.Interface())
		if err != nil {
			b.sendError(req, fmt.Errorf("marshal err: %v", err))
		}
		results[i] = string(marshal)
	}
	b.send(models.PacketAction, ActionResponse{
		ID:       req.ID,
		OK:       true,
		ErrorMsg: "",
		Data:     results,
	})
}

func (b *PluginBus) sendError(req ActionRequest, msg error) {
	log.Errorf("Error: %v", msg)
	b.send(models.PacketAction, ActionResponse{
		ID:       req.ID,
		OK:       false,
		ErrorMsg: msg.Error(),
		Data:     make([]string, 0),
	})
}

func (b *PluginBus) callFunc(name string, f any) {
	if f == nil || name == "" {
		panic("name or func must not be nil")
	}
	if reflect.TypeOf(f).Kind() != reflect.Func {
		panic("f must be a func")
	}
	_, loaded := b.call.LoadOrStore(name, f)
	if loaded {
		log.Errorf("call func %s already loaded", name)
	}
}

type (
	ToolFunc func(name string, f any)
	Tools    func(ToolFunc) string
)

func (b *PluginBus) BindTools(tools Tools) {
	toolsID := tools(b.callFunc)
	b.SendMeta(".tools", toolsID)
}

type ResourceFormatter = func(body string) (models.Resource, error)

func (b *PluginBus) BindResource(scheme string, provider models.ResourceProvider) ResourceFormatter {
	b.resource.RegisterResource(b.ID, scheme, provider)
	return func(body string) (models.Resource, error) {
		return models.Resource{
			PluginID: b.ID,
			Scheme:   scheme,
			Body:     body,
		}, nil
	}
}
