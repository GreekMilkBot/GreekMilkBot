package event

import (
	"encoding/json"

	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type EventType string

const (
	EventTypeMeta    EventType = "meta_event"
	EventTypeMessage EventType = "message"
	EventTypeNotice  EventType = "notice"
	EventTypeRequest EventType = "request"
)

func isValidEventType(eventType EventType) bool {
	switch eventType {
	case EventTypeMeta, EventTypeMessage, EventTypeNotice, EventTypeRequest:
		return true
	default:
		return false
	}
}

type Event interface {
	GetSelfId() int64
	GetType() EventType
}

type BaseEvent struct {
	Time     int64     `json:"time"`
	SelfID   int64     `json:"self_id"`
	PostType EventType `json:"post_type"`
}

func (e BaseEvent) GetSelfId() int64 {
	return e.SelfID
}

func (e BaseEvent) GetType() EventType {
	if !isValidEventType(e.PostType) {
		log.Error("Invalid event type: %s", e.PostType)
		return ""
	}
	return e.PostType
}

func JsonMsgToEvent(jsonData []byte) (Event, error) {
	var base BaseEvent
	if err := json.Unmarshal(jsonData, &base); err != nil {
		return nil, err
	}

	switch base.PostType {
	case "meta_event":
		return getMetaEvent(jsonData)
	case "message":
		var event MessageEvent
		return event, json.Unmarshal(jsonData, &event)
	case "notice":
		var event NoticeEvent
		return event, json.Unmarshal(jsonData, &event)
	case "request":
		var event RequestEvent
		return event, json.Unmarshal(jsonData, &event)
	default:
		return base, nil
	}
}
