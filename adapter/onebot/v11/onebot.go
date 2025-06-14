package v11

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/internal/api"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/internal/event"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/internal/models"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/driver"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

type OneBotV11Adapter struct {
	driver  driver.Driver
	actions *api.OneBotV11Actions

	selfId string
}

func NewOneBotV11Adapter(driver driver.Driver) *OneBotV11Adapter {
	return &OneBotV11Adapter{
		driver: driver,
	}
}

func (a *OneBotV11Adapter) Bind(ctx *bot.Bus) error {
	a.actions = api.NewOneBotV11Actions(a.driver.Send)
	a.bindFunc(ctx)
	return a.driver.Connect(ctx, a.handleMessage(ctx))
}

func (a *OneBotV11Adapter) handleMessage(ctx *bot.Bus) driver.Handler {
	return func(d driver.Driver, msg []byte) {
		log.Debug("OneBotV11Adapter: Received message: %s", msg)
		go func(m []byte) {
			if err := a.processMessage(ctx, m); err != nil {
				log.Error("OneBotV11Adapter: Failed to process message: %s", err)
			}
		}(msg)
	}
}

func (a *OneBotV11Adapter) processMessage(ctx *bot.Bus, msg []byte) error {
	e, err := event.JsonMsgToEvent(msg)
	if err != nil {
		return err
	}
	// init bot
	if a.selfId == "" {
		lce, ok := e.(event.MetaEventLifeCycle)
		if !ok {
			return nil
		}
		dt := a.driver.GetDriverType()
		if (dt == driver.DriverTypeWebSocketReverse || dt == driver.DriverTypeWebSocket) && lce.SubType != event.LifeCycleSubTypeConnect {
			log.Warn("OneBotV11Adapter: Unexpected life cycle event sub type: %s for ws driver", lce.SubType)
			return nil
		}
		if dt == driver.DriverTypeHTTPPost && (lce.SubType == event.LifeCycleSubTypeEnable || lce.SubType == event.LifeCycleSubTypeDisable) {
			log.Warn("OneBotV11Adapter: Unexpected life cycle event sub type: %s for http post driver", lce.SubType)
			return nil
		}
		a.selfId = fmt.Sprintf("%d", lce.SelfID)
		log.Info("OneBotV11Adapter: Bot initialized, self ID: %d", lce.SelfID)
		return nil
	}
	if message, ok := e.(event.ActionEvent); ok {
		if message.Code != 0 {
			a.actions.CallPacket(message.Echo, api.ActionState{
				Code:    message.Code,
				Message: message.Message,
				Data:    "",
			})
		} else {
			a.actions.CallPacket(message.Echo, api.ActionState{
				Code:    message.Code,
				Message: "",
				Data:    message.Data,
			})
		}
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
		ctx.SendMessage(*cMsg)
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
		Content: make(bot.Contents, 0),
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
				id := first.MsgData["id"].(string)
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
			case "at":
				var user *bot.User
				qq := message.MsgData["qq"].(string)
				if qq == "all" {
					qq = "*"
				} else {
					var info *api.StrangerInfo
					info, err = a.actions.GetStrangerInfo(qq)
					if err != nil {
						return nil, err
					}
					user = &bot.User{
						Id:     fmt.Sprintf("%d", info.UserID),
						Name:   info.Nickname,
						Avatar: fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=256", info.UserID),
					}
				}
				msg.Content = append(msg.Content, bot.ContentAt{Uid: qq, User: user})
			case "image":
				msg.Content = append(msg.Content, bot.ContentImage{
					URL:     message.MsgData["url"].(string),
					Summary: message.MsgData["summary"].(string),
				})
			case "face":
				msg.Content = append(msg.Content, bot.ContentUnknown{Type: "qq_face", Value: message.MsgData["id"].(string)})
			}
		}
	}
	return msg, nil
}

func (a *OneBotV11Adapter) bindFunc(ctx *bot.Bus) {
	ctx.CallFunc("send_private_msg", a.sendPrivateMessage)
	ctx.CallFunc("send_group_msg", a.sendGroupMessage)
}

func (a *OneBotV11Adapter) sendPrivateMessage(userId string, msg *bot.Contents) (string, error) {
	return a.sendMessage(userId, "", msg)
}

func (a *OneBotV11Adapter) sendGroupMessage(groupID string, msg *bot.Contents) (string, error) {
	return a.sendMessage("", groupID, msg)
}

func (a *OneBotV11Adapter) sendMessage(userId string, groupId string, msg *bot.Contents) (string, error) {
	var uid, gid uint64
	if userId != "" {
		i, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			return "", err
		}
		uid = uint64(i)
	}
	if groupId != "" {
		i, err := strconv.ParseInt(groupId, 10, 64)
		if err != nil {
			return "", err
		}
		gid = uint64(i)
	}
	message := make([]models.Message, 0)
	for _, content := range *msg {
		switch content.(type) {
		case bot.ContentText:
			message = append(message, models.Message{
				MsgType: "text",
				MsgData: map[string]interface{}{
					"text": content.(bot.ContentText).Text,
				},
			})
		case bot.ContentAt:
			u := content.(bot.ContentAt).Uid
			if u == "*" {
				u = "all"
			}
			message = append(message, models.Message{
				MsgType: "at",
				MsgData: map[string]interface{}{
					"qq": u,
				},
			})
		case bot.ContentImage:
			img := content.(bot.ContentImage)
			message = append(message, models.Message{
				MsgType: "image",
				MsgData: map[string]interface{}{
					"file": img.URL,
				},
			})
		}
	}
	sendMsg, err := a.actions.SendMsg(uid, gid, message)
	return fmt.Sprintf("%d", sendMsg), err
}
