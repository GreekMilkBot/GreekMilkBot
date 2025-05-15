package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type Adapter interface {
	Run(ctx *Bus) error
}

type BaseAdapter struct {
	Driver driver.Driver
	Bot    *bot.Bot
}

type Bus struct {
	ID string

	Tx chan<- bot.Packet
	Rx chan bot.ActionRequest

	context.Context

	call *sync.Map
}

func NewBus(pid string, bootCtx context.Context, tx chan bot.Packet) *Bus {
	bus := Bus{
		ID:      pid,
		Context: bootCtx,
		Tx:      tx,
		Rx:      make(chan bot.ActionRequest, 100),
		call:    &sync.Map{},
	}
	go bus.receiveLoop()
	return &bus
}

func (b *Bus) SendMessage(message bot.Message) {
	b.Tx <- bot.Packet{
		Plugin: b.ID,
		Type:   bot.PacketMessage,
		Data:   message,
	}
}

func (b *Bus) receiveLoop() {
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

func (b *Bus) exec(req bot.ActionRequest, value any) {
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
	b.Tx <- bot.Packet{
		Plugin: b.ID,
		Type:   bot.PacketAction,
		Data: bot.ActionResponse{
			ID:       req.ID,
			OK:       true,
			ErrorMsg: "",
			Data:     results,
		},
	}
}

func (b *Bus) sendError(req bot.ActionRequest, msg error) {
	log.Error("Error: %v", msg)
	b.Tx <- bot.Packet{
		Plugin: b.ID,
		Type:   bot.PacketAction,
		Data: bot.ActionResponse{
			ID:       req.ID,
			OK:       false,
			ErrorMsg: msg.Error(),
			Data:     make([]string, 0),
		},
	}
}

func (b *Bus) BindCall(name string, f any) {
	if f == nil || name == "" {
		panic("name or func must not be nil")
	}
	if reflect.TypeOf(f).Kind() != reflect.Func {
		panic("f must be a func")
	}
	b.call.Store(name, f)
}
