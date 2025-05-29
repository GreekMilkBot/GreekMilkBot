package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"sync"
	"sync/atomic"
	"time"
)

type OneBotV11Actions struct {
	driver driver.Driver
	result sync.Map
	index  atomic.Uint64
}

func NewOneBotV11Actions(d driver.Driver) (*OneBotV11Actions, error) {
	return &OneBotV11Actions{
		driver: d,
		result: sync.Map{},
		index:  atomic.Uint64{},
	}, nil
}

func (o *OneBotV11Actions) addHook(api string, args any, timeout time.Duration) (string, error) {
	id := fmt.Sprintf("%s_%d", api, o.index.Add(1))
	r := make(chan string)
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
	if err = o.driver.Send(string(req)); err != nil {
		return "", err
	}
	select {
	case data := <-r:
		return data, nil
	case <-time.After(timeout):
		return "", context.DeadlineExceeded
	}
}

func (o *OneBotV11Actions) CallPacket(id, data string) {
	value, loaded := o.result.LoadAndDelete(id)
	if !loaded {
		log.Warn("drop action before put %v", data)
		return
	}
	c := value.(chan string)
	c <- data
	close(c)
}
