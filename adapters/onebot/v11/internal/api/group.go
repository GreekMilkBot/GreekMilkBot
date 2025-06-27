package api

import (
	"encoding/json"
	"time"
)

type GroupInfo struct {
	GroupID        int64  `json:"group_id"`
	GroupName      string `json:"group_name"`
	MemberCount    int    `json:"member_count"`
	MAXMemberCount int    `json:"max_member_count"`
}

func (o *OneBotV11Actions) GetGroupInfo(groupID uint64) (*GroupInfo, error) {
	hook, err := o.addHook("get_group_info", map[string]any{
		"group_id": groupID,
	}, 1*time.Second)
	if err != nil {
		return nil, err
	}
	var result GroupInfo
	return &result, json.Unmarshal([]byte(hook), &result)
}
