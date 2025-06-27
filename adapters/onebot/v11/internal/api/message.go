package api

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/adapters/onebot/v11/internal/models"
)

func (o *OneBotV11Actions) GetMsg(msgId string) (*models.CommonMessage, error) {
	hook, err := o.addHook("get_msg", map[string]any{
		"message_id": msgId,
	}, 1*time.Second)
	if err != nil {
		return nil, err
	}
	var result models.CommonMessage
	return &result, json.Unmarshal([]byte(hook), &result)
}

func (o *OneBotV11Actions) SendMsg(userID, groupID uint64, message []models.Message) (uint, error) {
	args := map[string]any{
		"message": message,
	}
	if userID > 0 && groupID > 0 {
		return 0, errors.New("user and group cannot be set at the same time")
	}
	if userID > 0 {
		args["user_id"] = userID
	}
	if groupID > 0 {
		args["group_id"] = groupID
	}
	hook, err := o.addHook("send_msg", args, 1*time.Second)
	if err != nil {
		return 0, err
	}

	result := struct {
		MessageId uint `json:"message_id"`
	}{}
	return result.MessageId, json.Unmarshal([]byte(hook), &result)
}
