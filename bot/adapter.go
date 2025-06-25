package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type Adapter interface {
	Bind(ctx *Bus) error
}

type Bus struct {
	ID string

	tx chan<- Packet
	rx chan ActionRequest

	context.Context

	call *sync.Map
}

func NewBus(id string, ctx context.Context, tx chan Packet) *Bus {
	bus := Bus{
		ID:      id,
		Context: ctx,
		tx:      tx,
		rx:      make(chan ActionRequest, 100),
		call:    &sync.Map{},
	}
	go bus.receiveLoop()
	return &bus
}

func (b *Bus) SendMessage(message Message) {
	b.tx <- Packet{
		Plugin: b.ID,
		Type:   PacketMessage,
		Data:   message,
	}
}

func (b *Bus) SendEvent(event Event) {
	b.tx <- Packet{
		Plugin: b.ID,
		Type:   PacketAction,
		Data:   event,
	}
}

func (b *Bus) SendMeta(key string, value string) {
	b.tx <- Packet{
		Plugin: b.ID,
		Type:   PacketMeta,
		Data: Meta{
			Key:   key,
			Value: value,
		},
	}
}

func (b *Bus) NewRequest(req ActionRequest) {
	b.rx <- req
}

func (b *Bus) receiveLoop() {
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
			b.sendError(req, v.Interface().(error))
			return
		}
		marshal, err := json.Marshal(v.Interface())
		if err != nil {
			b.sendError(req, fmt.Errorf("marshal err: %v", err))
		}
		results[i] = string(marshal)
	}
	b.tx <- Packet{
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
	log.Errorf("Error: %v", msg)
	b.tx <- Packet{
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
	_, loaded := b.call.LoadOrStore(name, f)
	if loaded {
		log.Errorf("call func %s already loaded", name)
	}
}

type Sender interface {
	SendPrivateMessage(userId string, msg *ClientMessage) (string, error)
	SendGroupMessage(groupID string, msg *ClientMessage) (string, error)
}

func (b *Bus) SendBinding(s Sender) {
	b.CallFunc("send_private_msg", s.SendPrivateMessage)
	b.CallFunc("send_group_msg", s.SendGroupMessage)
}
