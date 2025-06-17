package gmb

import (
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/bot"
)

type ClientBus struct {
	call ClientCallHandle
	id   string
}

func NewClientBus(id string, call ClientCallHandle) *ClientBus {
	return &ClientBus{
		id:   id,
		call: call,
	}
}

func (b *ClientBus) SendMessage(receive *bot.Message, message *bot.Contents) (string, error) {
	if receive.Guild != nil {
		return b.SendGroupMessage(receive.Guild.Id, message)
	} else {
		return b.SendPrivateMessage(receive.Owner.Id, message)
	}
}

func (b *ClientBus) SendPrivateMessage(userID string, message *bot.Contents) (string, error) {
	var messageId string
	return messageId, b.call(b.id, "send_private_msg", []any{userID, message}, []any{&messageId}, 1*time.Second)
}

func (b *ClientBus) SendGroupMessage(groupId string, message *bot.Contents) (string, error) {
	var messageId string
	return messageId, b.call(b.id, "send_group_msg", []any{groupId, message}, []any{&messageId}, 1*time.Second)
}
