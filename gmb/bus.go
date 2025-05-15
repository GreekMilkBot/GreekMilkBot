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
)

type ClientBus struct {
	pluginID string
	tx       chan<- bot.Packet
	call     *sync.Map
}

func (b *ClientBus) Call(key string, params []any, result []any, timeout time.Duration) error {
	id := fmt.Sprintf("%s-%s-%d-%.2f", b.pluginID, key, time.Now().Unix(), rand.Float64())
	for _, item := range result {
		paramType := reflect.TypeOf(item)
		if paramType.Kind() != reflect.Ptr {
			return fmt.Errorf("result must be a pointer")
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
	b.call.Store(id, resultChan)
	b.tx <- bot.Packet{
		Plugin: b.pluginID,
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
		if value, loaded := b.call.LoadAndDelete(id); loaded {
			close(value.(chan bot.ActionResponse))
		}
		return context.DeadlineExceeded
	}
	return nil
}

func (b *ClientBus) SendMessage(receive *bot.Message, message bot.Contents) (string, error) {
	if receive.Guild != nil {
		return b.SendGroupMessage(receive.Guild.Id, message)
	} else {
		return b.SendPrivateMessage(receive.Owner.Id, message)
	}
}

func (b *ClientBus) SendPrivateMessage(userID string, message bot.Contents) (string, error) {
	var messageId string
	contents, err := message.ToRAWContents()
	if err != nil {
		return "", err
	}
	return messageId, b.Call("send_private_msg", []any{userID, contents}, []any{&messageId}, 1*time.Second)
}

func (b *ClientBus) SendGroupMessage(groupId string, message bot.Contents) (string, error) {
	var messageId string
	contents, err := message.ToRAWContents()
	if err != nil {
		return "", err
	}
	return messageId, b.Call("send_group_msg", []any{groupId, contents}, []any{&messageId}, 1*time.Second)
}
