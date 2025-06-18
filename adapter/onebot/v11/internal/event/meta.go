package event

import (
	"encoding/json"
	"fmt"
)

// https://github.com/botuniverse/onebot-11/blob/master/event/meta.md

type MetaEventType string

const (
	MetaEventTypeLifeCycle MetaEventType = "lifecycle"
	MetaEventTypeHeartbeat MetaEventType = "heartbeat"
)

func IsValidMetaEventType(t MetaEventType) bool {
	switch t {
	case MetaEventTypeLifeCycle, MetaEventTypeHeartbeat:
		return true
	default:
		return false
	}
}

type MetaEvent struct {
	BaseEvent
	MetaEventType MetaEventType `json:"meta_event_type"`
}

func getMetaEvent(jsonData []byte) (Event, error) {
	var event MetaEvent
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return event, err
	}

	switch event.MetaEventType {
	case MetaEventTypeLifeCycle:
		var e MetaEventLifeCycle
		return e, json.Unmarshal(jsonData, &e)
	case MetaEventTypeHeartbeat:
		var e MetaEventHeartbeat
		return e, json.Unmarshal(jsonData, &e)
	default:
		return event, fmt.Errorf("invalid meta event type: %s", event.MetaEventType)
	}
}

type LifeCycleSubType string

const (
	LifeCycleSubTypeEnable  LifeCycleSubType = "enable"
	LifeCycleSubTypeDisable LifeCycleSubType = "disable"
	LifeCycleSubTypeConnect LifeCycleSubType = "connect"
)

type MetaEventLifeCycle struct {
	MetaEvent
	SubType LifeCycleSubType `json:"sub_type"`
}

type OneBotStatus struct {
	Online bool `json:"online"`
	Good   bool `json:"good"`
}

type MetaEventHeartbeat struct {
	MetaEvent
	Status   OneBotStatus `json:"status"`
	Interval int64        `json:"interval"`
}
