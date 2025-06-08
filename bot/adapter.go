package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/GreekMilkBot/GreekMilkBot/driver"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type Adapter interface {
	Bind(ctx *Bus) error
}

type BaseAdapter struct {
	Driver driver.Driver
	Booted *sync.WaitGroup
	Bot    *Bot
}

func NewBaseAdapter(driver driver.Driver) BaseAdapter {
	return BaseAdapter{
		Driver: driver,
		Booted: new(sync.WaitGroup),
	}
}

type Bus struct {
	ID string

	Tx chan<- Packet
	Rx chan ActionRequest

	context.Context

	call *sync.Map
}

func NewBus(id string, ctx context.Context, tx chan Packet) *Bus {
	bus := Bus{
		ID:      id,
		Context: ctx,
		Tx:      tx,
		Rx:      make(chan ActionRequest, 100),
		call:    &sync.Map{},
	}
	go bus.receiveLoop()
	return &bus
}

func (b *Bus) SendMessage(message Message) {
	b.Tx <- Packet{
		Plugin: b.ID,
		Type:   PacketMessage,
		Data:   message,
	}
}

func (b *Bus) receiveLoop() {
	defer close(b.Rx)
	for {
		select {
		case <-b.Context.Done():
			return
		case req := <-b.Rx:
			if value, ok := b.call.Load(req.Action); ok {
				go b.exec(req, value)
			} else {
				b.sendError(req, errors.New("func not found"))
			}
		}
	}
}

func (b *Bus) exec(req ActionRequest, value any) {
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
			b.sendError(req, fmt.Errorf(v.Interface().(error).Error()))
			return
		}
		marshal, err := json.Marshal(v.Interface())
		if err != nil {
			b.sendError(req, fmt.Errorf("marshal err: %v", err))
		}
		results[i] = string(marshal)
	}
	b.Tx <- Packet{
		Plugin: b.ID,
		Type:   PacketAction,
		Data: ActionResponse{
			ID:       req.ID,
			OK:       true,
			ErrorMsg: "",
			Data:     results,
		},
	}
}

func (b *Bus) sendError(req ActionRequest, msg error) {
	log.Error("Error: %v", msg)
	b.Tx <- Packet{
		Plugin: b.ID,
		Type:   PacketAction,
		Data: ActionResponse{
			ID:       req.ID,
			OK:       false,
			ErrorMsg: msg.Error(),
			Data:     make([]string, 0),
		},
	}
}

func (b *Bus) CallFunc(name string, f any) {
	if f == nil || name == "" {
		panic("name or func must not be nil")
	}
	if reflect.TypeOf(f).Kind() != reflect.Func {
		panic("f must be a func")
	}
	b.call.Store(name, f)
}
