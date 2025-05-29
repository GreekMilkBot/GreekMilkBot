package v11

import (
	"fmt"
	"github.com/GreekMilkBot/GreekMilkBot/adapter"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/api"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/event"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/models"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"sync"
	"time"
)

type OneBotV11Adapter struct {
	adapter.BaseAdapter

	eventBus adapter.Bus

	handler sync.Map
	actions *api.OneBotV11Actions
}

func NewOneBotV11Adapter(driver driver.Driver) *OneBotV11Adapter {
	return &OneBotV11Adapter{
		BaseAdapter: adapter.BaseAdapter{
			Driver: driver,
		},
		handler: sync.Map{},
	}
}

func (a *OneBotV11Adapter) Run(ctx adapter.Bus) error {
	err := a.Driver.Connect(ctx)
	a.eventBus = ctx
	if err != nil {
		log.Error("OneBotV11Adapter.Run:%s", err)
		return err
	}
	a.actions, err = api.NewOneBotV11Actions(a.Driver)
	if err != nil {
		log.Error("OneBotV11Adapter.Run:%s", err)
		return err
	}
	a.Driver.SetReceiveHandler(a.handleMessage)
	return nil
}

func (a *OneBotV11Adapter) handleMessage(d driver.Driver, msg []byte) {
	log.Debug("OneBotV11Adapter: Received message: %s", msg)
	go func(m []byte) {
		if err := a.processMessage(d, m); err != nil {
			log.Error("OneBotV11Adapter: Failed to process message: %s", err)
		}
	}(msg)
}

func (a *OneBotV11Adapter) processMessage(d driver.Driver, msg []byte) error {
	e, err := event.JsonMsgToEvent(msg)
	if err != nil {
		return err
	}
	// init bot
	if a.Bot == nil {
		lce, ok := e.(event.MetaEventLifeCycle)
		if !ok {
			return nil
		}
		dt := d.GetDriverType()
		if (dt == driver.DriverTypeWebSocketReverse || dt == driver.DriverTypeWebSocket) && lce.SubType != event.LifeCycleSubTypeConnect {
			log.Warn("OneBotV11Adapter: Unexpected life cycle event sub type: %s for ws driver", lce.SubType)
			return nil
		}
		if dt == driver.DriverTypeHTTPPost && (lce.SubType == event.LifeCycleSubTypeEnable || lce.SubType == event.LifeCycleSubTypeDisable) {
			log.Warn("OneBotV11Adapter: Unexpected life cycle event sub type: %s for http post driver", lce.SubType)
			return nil
		}
		a.Bot = bot.NewBot(fmt.Sprintf("%d", lce.SelfID))
		log.Info("OneBotV11Adapter: Bot initialized, self ID: %s", lce.SelfID)
		return nil
	}
	if message, ok := e.(event.ActionEvent); ok {
		a.actions.CallPacket(message.Echo, message.Data)
	}
	if message, ok := e.(event.MessageEvent); ok {
		if message.UserId == message.SelfID {
			log.Debug("OneBotV11Adapter: skip self message: %s", message)
			return nil
		}
		cMsg, err := a.covertMessage(&message.CommonMessage, 5)
		if err != nil {
			log.Error("OneBotV11Adapter: Failed to covertMessage: %s", err)
			return nil
		}
		_ = a.eventBus.SendMessage(*cMsg)
	}
	return nil
}

func (a *OneBotV11Adapter) covertMessage(e *models.CommonMessage, depth int) (*bot.Message, error) {
	msg := &bot.Message{
		ID: fmt.Sprintf("%d", e.MessageID),
		Owner: &bot.User{
			Id:     fmt.Sprintf("%d", e.Sender.UserId),
			Name:   e.Sender.Nickname,
			Avatar: fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=256", e.Sender.UserId),
		},
		Guild:   nil,
		Quote:   nil,
		Created: time.Unix(e.Time, 0),
		Updated: time.Unix(e.Time, 0),
	}
	if e.MessageType == models.MessageTypeGroup && e.GroupID != 0 {
		info, err := a.actions.GetGroupInfo(e.GroupID)
		if err != nil {
			return nil, err
		}
		msg.MsgType = "guild"
		msg.Guild = &bot.Guild{
			Id:     fmt.Sprintf("%d", info.GroupID),
			Name:   info.GroupName,
			Avatar: fmt.Sprintf("https://p.qlogo.cn/gh/%d/%d/640", info.GroupID, info.GroupID),
		}
		if first := e.Message[0]; first.MsgType == "reply" {
			e.Message = e.Message[1:]
			if depth > 0 {
				id := first.MsgData["id"].(int)
				getMsg, err := a.actions.GetMsg(id)
				if err != nil {
					return nil, err
				}
				depth--
				msg.Quote, err = a.covertMessage(getMsg, depth)
				if err != nil {
					return nil, err
				}
			}
		}
		for _, message := range e.Message {
			switch message.MsgType {
			case "text":
				msg.Content = append(msg.Content, bot.ContentText{Text: message.MsgData["text"].(string)})
			}
		}
	}
	return msg, nil
}
