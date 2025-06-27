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

type AdapterBus struct {
	ID string // bot id

	tx chan<- models.Packet
	rx chan ActionRequest

	context.Context

	call *sync.Map
}

func NewAdapterBus(id string, ctx context.Context, tx chan models.Packet) *AdapterBus {
	bus := AdapterBus{
		ID:      id,
		Context: ctx,
		tx:      tx,
		rx:      make(chan ActionRequest, 100),
		call:    &sync.Map{},
	}
	go bus.receiveLoop()
	return &bus
}

func (b *AdapterBus) SendMessage(message models.Message) {
	b.tx <- models.Packet{
		Plugin: b.ID,
		Type:   models.PacketMessage,
		Data:   message,
	}
}

func (b *AdapterBus) SendEvent(event models.Event) {
	b.tx <- models.Packet{
		Plugin: b.ID,
		Type:   models.PacketAction,
		Data:   event,
	}
}

func (b *AdapterBus) SendMeta(key string, value string) {
	b.tx <- models.Packet{
		Plugin: b.ID,
		Type:   models.PacketMeta,
		Data: models.Meta{
			Key:   key,
			Value: value,
		},
	}
}

func (b *AdapterBus) NewRequest(req ActionRequest) {
	b.rx <- req
}

func (b *AdapterBus) receiveLoop() {
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

func (b *AdapterBus) exec(req ActionRequest, value any) {
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
	b.tx <- models.Packet{
		Plugin: b.ID,
		Type:   models.PacketAction,
		Data: ActionResponse{
			ID:       req.ID,
			OK:       true,
			ErrorMsg: "",
			Data:     results,
		},
	}
}

func (b *AdapterBus) sendError(req ActionRequest, msg error) {
	log.Errorf("Error: %v", msg)
	b.tx <- models.Packet{
		Plugin: b.ID,
		Type:   models.PacketAction,
		Data: ActionResponse{
			ID:       req.ID,
			OK:       false,
			ErrorMsg: msg.Error(),
			Data:     make([]string, 0),
		},
	}
}

func (b *AdapterBus) CallFunc(name string, f any) {
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

func (b *AdapterBus) BindTools(tools Tools) {
	toolsID := tools(b.CallFunc)
	b.SendMeta(".tools", toolsID)
}
