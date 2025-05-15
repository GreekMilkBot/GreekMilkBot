package event

import (
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/internal/models"
)

// https://github.com/botuniverse/onebot-11/blob/master/event/message.md

type MessageEvent struct {
	BaseEvent
	models.CommonMessage
}
