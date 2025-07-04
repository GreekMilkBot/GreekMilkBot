package event

import (
	"encoding/json"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/log"
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
	GetSelfId() uint64
	GetType() EventType
}

type BaseEvent struct {
	SelfID   uint64    `json:"self_id"`
	PostType EventType `json:"post_type"`

	Echo string `json:"echo,omitempty"`
}

func (e BaseEvent) GetSelfId() uint64 {
	return e.SelfID
}

func (e BaseEvent) GetType() EventType {
	if !isValidEventType(e.PostType) {
		log.Errorf("Invalid event type: %s", e.PostType)
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
		if base.Echo != "" {
			var event ActionEvent
			return event, json.Unmarshal(jsonData, &event)
		}
		return base, nil
	}
}
