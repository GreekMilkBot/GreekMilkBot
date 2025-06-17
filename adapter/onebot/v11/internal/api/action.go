package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type OneBotV11Actions struct {
	sender func(string) error
	result sync.Map
	index  atomic.Uint64
}

func NewOneBotV11Actions(sender func(string) error) *OneBotV11Actions {
	return &OneBotV11Actions{
		sender: sender,
		result: sync.Map{},
		index:  atomic.Uint64{},
	}
}

func (o *OneBotV11Actions) addHook(api string, args any, timeout time.Duration) (string, error) {
	id := fmt.Sprintf("%s_%d", api, o.index.Add(1))
	r := make(chan ActionState)
	o.result.Store(id, r)
	defer func() {
		if _, ok := o.result.LoadAndDelete(id); ok {
			close(r)
		}
	}()
	req, err := json.Marshal(map[string]interface{}{
		"echo":   id,
		"action": api,
		"params": args,
	})
	if err != nil {
		return "", err
	}
	if err = o.sender(string(req)); err != nil {
		return "", err
	}
	select {
	case data := <-r:
		if data.Code == 0 {
			return data.Data, nil
		} else {
			return "", errors.New(data.Message)
		}
	case <-time.After(timeout):
		return "", context.DeadlineExceeded
	}
}

type ActionState struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (o *OneBotV11Actions) CallPacket(id string, data ActionState) {
	value, loaded := o.result.LoadAndDelete(id)
	if !loaded {
		log.Warnf("drop action before put %v", data)
		return
	}
	c := value.(chan ActionState)
	c <- data
	close(c)
}
