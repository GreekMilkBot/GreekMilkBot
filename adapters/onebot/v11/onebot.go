package v11

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/adapters/onebot/v11/internal/utils"

	"github.com/GreekMilkBot/GreekMilkBot/adapters/onebot/v11/apis"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/tools"

	bot_models "github.com/GreekMilkBot/GreekMilkBot/pkg/models"

	"github.com/GreekMilkBot/GreekMilkBot/adapters/onebot/v11/internal/api"
	"github.com/GreekMilkBot/GreekMilkBot/adapters/onebot/v11/internal/driver"
	"github.com/GreekMilkBot/GreekMilkBot/adapters/onebot/v11/internal/event"
	"github.com/GreekMilkBot/GreekMilkBot/adapters/onebot/v11/internal/models"
	gmbcore "github.com/GreekMilkBot/GreekMilkBot/pkg"
	"github.com/GreekMilkBot/GreekMilkBot/pkg/log"
)

func init() {
	gmbcore.RegisterAdapter("onebot11", func(ctx context.Context, url url.URL) (gmbcore.Plugin, error) {
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

	bind           *atomic.Bool
	driver         *driver.WebSocketDriver
	ctx            *gmbcore.PluginBus
	imageFormatter func(body string) bot_models.Resource
}

func (a *OneBotV11Adapter) Metadata(scheme, body string) (*bot_models.Metadata, error) {
	if scheme != "image" {
		return nil, errors.New("unsupported scheme :" + scheme)
	}
	return bot_models.HttpMetadata(utils.OneBotHttpClient, body)
}

func (a *OneBotV11Adapter) Reader(scheme, body string) (io.ReadCloser, error) {
	if scheme != "image" {
		return nil, errors.New("unsupported scheme :" + scheme)
	}
	return bot_models.HttpReader(utils.OneBotHttpClient, body)
}

func NewOneBotV11Adapter(driver *driver.WebSocketDriver) *OneBotV11Adapter {
	return &OneBotV11Adapter{
		bind:    new(atomic.Bool),
		driver:  driver,
		selfId:  new(atomic.Uint64),
		actions: api.NewOneBotV11Actions(driver.Send),
	}
}

func (a *OneBotV11Adapter) Bind(ctx *gmbcore.PluginBus) error {
	if a.bind.Swap(true) {
		return errors.New("already bind")
	}
	a.ctx = ctx
	imageFormatter := ctx.BindResource("image", a)
	a.imageFormatter = func(body string) bot_models.Resource {
		formatter, _ := imageFormatter(body)
		return formatter
	}
	ctx.BindTools(tools.Sender(a))
	return a.driver.Bind(func(msg []byte) {
		log.Debugf("OneBotV11Adapter: Received message: %s", msg)
		go func(m []byte) {
			if err := a.processMessage(ctx, m); err != nil {
				log.Errorf("OneBotV11Adapter: Failed to process message: %s", err)
			}
		}(msg)
	})
}

func (a *OneBotV11Adapter) processMessage(ctx *gmbcore.PluginBus, msg []byte) error {
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
			log.Debugf("OneBotV11Adapter: skip self message: %v", message)
			return nil
		}
		cMsg, err := a.covertMessage(&message.CommonMessage, 5)
		if err != nil {
			log.Errorf("OneBotV11Adapter: Failed to covertMessage: %s", err)
			return nil
		}
		ctx.SendMessage(*cMsg)
	}
	if e, ok := e.(event.NoticeEvent); ok {
		ctx.SendEvent(bot_models.Event{
			Type: fmt.Sprintf("onebot11_%s", e.GetNoticeType()),
			Data: e,
		})
	}
	return nil
}

func (a *OneBotV11Adapter) covertMessage(e *models.CommonMessage, depth int) (*bot_models.Message, error) {
	msg := &bot_models.Message{
		ID: fmt.Sprintf("%d", e.MessageID),
		Owner: &bot_models.GuildMember{
			User: &bot_models.User{
				Id:     fmt.Sprintf("%d", e.Sender.UserId),
				Name:   e.Sender.Nickname,
				Avatar: a.imageFormatter(fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=256", e.Sender.UserId)),
			},
			GuildRole: make([]string, 0),
		},
		Content: make(bot_models.Contents, 0),
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
		msg.Guild = &bot_models.Guild{
			Id:     fmt.Sprintf("%d", info.GroupID),
			Name:   info.GroupName,
			Avatar: a.imageFormatter(fmt.Sprintf("https://p.qlogo.cn/gh/%d/%d/640", info.GroupID, info.GroupID)),
		}
		msg.Owner.GuildName = e.Sender.Card
		msg.Owner.GuildRole = []string{e.Sender.Role}
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
				last := len(msg.Content) - 1
				if last >= 0 {
					if f, ok := msg.Content[last].(bot_models.ContentText); ok {
						f.Text = f.Text + message.MsgData["text"].(string)
						msg.Content[last] = f
						continue
					}
				}
				msg.Content = append(msg.Content, bot_models.ContentText{Text: message.MsgData["text"].(string)})
			case "at":
				var user *bot_models.User
				qq := message.MsgData["qq"].(string)
				if qq == "all" {
					qq = "*"
				} else {
					var info *api.StrangerInfo
					info, err = a.actions.GetStrangerInfo(qq)
					if err != nil {
						return nil, err
					}
					user = &bot_models.User{
						Id:     fmt.Sprintf("%d", info.UserID),
						Name:   info.Nickname,
						Avatar: a.imageFormatter(fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=256", info.UserID)),
					}
				}
				msg.Content = append(msg.Content, bot_models.ContentAt{Uid: qq, User: user})
			case "image":
				msg.Content = append(msg.Content, bot_models.ContentImage{
					Resource: a.imageFormatter(message.MsgData["url"].(string)),
					Summary:  message.MsgData["summary"].(string),
				})
			default:
				rawMsg, _ := json.Marshal(message.MsgData)
				msg.Content = append(msg.Content, bot_models.ContentUnknown{
					Type:  fmt.Sprintf("onebot11_%s", message.MsgType),
					Value: string(rawMsg),
				})
			}
		}
	}
	return msg, nil
}

func (a *OneBotV11Adapter) SendPrivateMessage(userId string, msg *tools.SenderMessage) (string, error) {
	return a.sendMessage(userId, "", msg)
}

func (a *OneBotV11Adapter) SendGroupMessage(groupID string, msg *tools.SenderMessage) (string, error) {
	return a.sendMessage("", groupID, msg)
}

func (a *OneBotV11Adapter) sendMessage(userId string, groupId string, msg *tools.SenderMessage) (string, error) {
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
	if msg.QuoteID != "" {
		message = append(message, models.Message{
			MsgType: "reply",
			MsgData: map[string]interface{}{
				"id": msg.QuoteID,
			},
		})
	}
	for _, content := range *msg.Message {
		switch content := content.(type) {
		case bot_models.ContentText:
			last := len(message) - 1
			if last > 0 && message[last].MsgType == "text" {
				m := message[last]
				m.MsgData["text"] = message[last].MsgData["text"].(string) + content.Text
				message[last] = m
			} else {
				message = append(message, models.Message{
					MsgType: "text",
					MsgData: map[string]interface{}{
						"text": content.Text,
					},
				})
			}
		case bot_models.ContentAt:
			u := content.Uid
			if u == "*" {
				u = "all"
			}
			message = append(message, models.Message{
				MsgType: "at",
				MsgData: map[string]interface{}{
					"qq": u,
				},
			})
		case bot_models.ContentImage:
			img := content
			if img.Resource.PluginID == a.ctx.ID {
				message = append(message, models.Message{
					MsgType: "image",
					MsgData: map[string]interface{}{
						"file": img.Resource.Body,
					},
				})
			} else {
				blob, err := a.ctx.ResourceBlob(&img.Resource)
				if err != nil {
					return "", err
				}
				all, err := io.ReadAll(blob)
				_ = blob.Close()
				if err != nil {
					return "", err
				}
				message = append(message, models.Message{
					MsgType: "image",
					MsgData: map[string]interface{}{
						"file": fmt.Sprintf("base64://%s", base64.URLEncoding.EncodeToString(all)),
					},
				})
			}
		case apis.OneBotCustomContent:
			message = append(message, models.Message{
				MsgType: content.Type,
				MsgData: content.Data,
			})
		}
	}
	sendMsg, err := a.actions.SendMsg(uid, gid, message)
	return fmt.Sprintf("%d", sendMsg), err
}
