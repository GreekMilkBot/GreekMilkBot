package api

import (
	"encoding/json"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/models"
	"time"
)

func (o *OneBotV11Actions) GetMsg(msgId int) (*models.CommonMessage, error) {
	hook, err := o.addHook("get_msg", map[string]any{
		"message_id": msgId,
	}, 1*time.Second)
	if err != nil {
		return nil, err
	}
	var result models.CommonMessage
	return &result, json.Unmarshal([]byte(hook), &result)
}
