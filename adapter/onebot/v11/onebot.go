package v11

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/gmb"

	"github.com/GreekMilkBot/GreekMilkBot/driver"

	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/internal/api"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/internal/event"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/onebot/v11/internal/models"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

func init() {
	gmb.RegisterAdapter("onebot11", func(ctx context.Context, url url.URL) (bot.Adapter, error) {
		if url.Scheme != "ws" && url.Scheme != "wss" {
			return nil, errors.New("unsupported scheme :" + url.Scheme)
		}
		token := url.Query().Get("token")
		retry := url.Query().Get("retry") == "true"
		return NewOneBotV11Adapter(
			driver.NewWebSocketDriver(ctx, fmt.Sprintf("%s://%s%s", url.Scheme, url.Host, url.Path),
				token, retry)), nil
	})
}

type OneBotV11Adapter struct {
	actions *api.OneBotV11Actions
	selfId  *atomic.Uint64

	bind   *atomic.Bool
	driver *driver.WebSocketDriver
}

func NewOneBotV11Adapter(driver *driver.WebSocketDriver) *OneBotV11Adapter {
	return &OneBotV11Adapter{
		bind:    new(atomic.Bool),
		driver:  driver,
		selfId:  new(atomic.Uint64),
		actions: api.NewOneBotV11Actions(driver.Send),
	}
}

func (a *OneBotV11Adapter) Bind(ctx *bot.Bus) error {
	if a.bind.Swap(true) {
		return errors.New("already bind")
	}
	ctx.CallFunc("send_private_msg", a.sendPrivateMessage)
	ctx.CallFunc("send_group_msg", a.sendGroupMessage)
	return a.driver.Bind(func(msg []byte) {
		log.Debugf("OneBotV11Adapter: Received message: %s", msg)
		go func(m []byte) {
			if err := a.processMessage(ctx, m); err != nil {
				log.Errorf("OneBotV11Adapter: Failed to process message: %s", err)
			}
		}(msg)
	})
}

func (a *OneBotV11Adapter) processMessage(ctx *bot.Bus, msg []byte) error {
	e, err := event.JsonMsgToEvent(msg)
	if err != nil {
		return err
	}
	// init bot
	if a.selfId.Load() == 0 {
		lce, ok := e.(event.MetaEventLifeCycle)
		if !ok {
			return nil
		}
		a.selfId.Store(lce.SelfID)
		log.Infof("OneBotV11Adapter: Bot initialized, self ID: %d", lce.SelfID)
		ctx.SendMeta("id", fmt.Sprintf("%d", lce.SelfID))
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
			log.Debugf("OneBotV11Adapter: skip self message: %s", message)
			return nil
		}
		cMsg, err := a.covertMessage(&message.CommonMessage, 5)
		if err != nil {
			log.Errorf("OneBotV11Adapter: Failed to covertMessage: %s", err)
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
