package tools

import (
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"

	"github.com/GreekMilkBot/GreekMilkBot/pkg"

	gmb "github.com/GreekMilkBot/GreekMilkBot"
)

var SenderToolID = "bot_tools_sender"

type SenderClient gmb.BotContext

func (b SenderClient) SendMessage(receive *models.Message, message *SenderMessage) (string, error) {
	if receive.Guild != nil {
		return b.SendGroupMessage(receive.Guild.Id, message)
	} else {
		return b.SendPrivateMessage(receive.Owner.Id, message)
	}
}

func (b SenderClient) SendPrivateMessage(userID string, message *SenderMessage) (string, error) {
	var messageId string
	return messageId, b.Call(b.BotID, "send_private_msg", []any{userID, message}, []any{&messageId}, 1*time.Second)
}

func (b SenderClient) SendGroupMessage(groupId string, message *SenderMessage) (string, error) {
	var messageId string
	return messageId, b.Call(b.BotID, "send_group_msg", []any{groupId, message}, []any{&messageId}, 1*time.Second)
}

type SenderMessage struct {
	QuoteID string           `json:"quote_id"`
	Message *models.Contents `json:"contents"`
}

type SenderAdapter interface {
	SendPrivateMessage(userId string, msg *SenderMessage) (string, error)
	SendGroupMessage(groupID string, msg *SenderMessage) (string, error)
}

func Sender(msg SenderAdapter) core.Tools {
	return func(toolFunc core.ToolFunc) string {
		toolFunc("send_private_msg", msg.SendPrivateMessage)
		toolFunc("send_group_msg", msg.SendGroupMessage)
		return SenderToolID
	}
}
