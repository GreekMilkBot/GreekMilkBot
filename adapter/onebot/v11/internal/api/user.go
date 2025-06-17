package api

import (
	"encoding/json"
	"time"
)

type StrangerInfo struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
}

func (o *OneBotV11Actions) GetStrangerInfo(userID string) (*StrangerInfo, error) {
	hook, err := o.addHook("get_stranger_info", map[string]any{
		"user_id": userID,
	}, 1*time.Second)
	if err != nil {
		return nil, err
	}
	var result StrangerInfo
	return &result, json.Unmarshal([]byte(hook), &result)
}
