package dummy

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/GreekMilkBot/GreekMilkBot/adapter/dummy/internal"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/dummy/internal/server"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/dummy/static"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/gmb"
	"github.com/GreekMilkBot/GreekMilkBot/log"
)

//go:embed default.json
var defCfg []byte

func init() {
	gmb.RegisterAdapter("dummy", func(ctx context.Context, url url.URL) (bot.Adapter, error) {
		if url.Scheme != "bind" {
			return nil, errors.New("unsupported scheme :" + url.Scheme)
		}
		listen, err := net.Listen("tcp", url.Host)
		if err != nil {
			return nil, err
		}
		wrapper := internal.NewWrapper()
		include := url.Query().Get("include")
		if include != "" {
			var data []byte
			if include == "default" {
				log.Infof("当前使用默认测试数据")
				data = defCfg
			} else {
				data, err = os.ReadFile(include)
				if err != nil {
					return nil, err
				}
			}
			if err := json.Unmarshal(data, wrapper); err != nil {
				return nil, err
			}

		}
		wrapper.Handle("/", http.FileServerFS(static.FS))
		go func() {
			select {
			case <-ctx.Done():
				_ = listen.Close()
				_ = wrapper.Close()
			}
		}()
		go func() {
			_ = http.Serve(listen, wrapper)
		}()
		return &DummyAdapter{
			wrapper,
		}, nil
	})
}

type DummyAdapter struct {
	wrapper *internal.Wrapper
}

func (d *DummyAdapter) Bind(ctx *bot.Bus) error {
	d.wrapper.BindBotMessage = func(msg server.QueryMessageResp) {
		message, err := d.Dummy2Message(msg, 5)
		if err != nil {
			log.Warnf("message covert error %s", err.Error())
			return
		}
		ctx.SendMessage(*message)
	}
	ctx.SendBinding(d)
	return nil
}

func (d *DummyAdapter) Dummy2Message(msg server.QueryMessageResp, depth int) (*bot.Message, error) {
	if depth == 0 {
		return nil, errors.New("depth is zero")
	}
	content := msg.Content
	depth = depth - 1
	result := bot.Message{
		ID: content.ID,
		Owner: &bot.GuildMember{
			User: &bot.User{
				Id:     content.Sender.ID,
				Name:   content.Sender.Name,
				Avatar: content.Sender.Avatar,
			},
			GuildRole: make([]string, 0),
		},
		Created: content.CreateAt,
		Updated: content.CreateAt,
	}
	if msg.Type == "group" {
		result.MsgType = "guild"
		result.Guild = &bot.Guild{
			Id:     msg.TargetID,
			Name:   msg.TargetName,
			Avatar: msg.TargetAvatar,
		}
		result.Owner.GuildName = content.Sender.AliasName
	}
	if content.ReferID != "" {
		query, err := d.wrapper.Server.QueryMessage(content.ReferID)
		if err != nil {
			return nil, err
		}
		botMsg, err := d.Dummy2Message(*query, depth)
		if err != nil {
			return nil, err
		}
		if len(botMsg.Content) >= 0 {
			// 有些内容无法识别导致内容为空
			result.Quote = botMsg
		}
	}
	for _, content := range msg.Content.Content {
		switch content.Type {
		case "text":
			result.Content = append(result.Content, bot.ContentText{Text: content.Data})
		case "image":
			if strings.HasPrefix(content.Data, "data:") {
				image, err := NewDataUriContentImage(content.Data)
				if err != nil {
					return nil, err
				}
				result.Content = append(result.Content, image)
			} else {
				result.Content = append(result.Content, bot.ContentImage{
					URL:     content.Data,
					Summary: "",
				})
			}
		case "at":
			u, err := d.wrapper.Server.GetUser(content.Data)
			if err != nil {
				return nil, err
			}
			result.Content = append(result.Content, bot.ContentAt{
				Uid: content.Data,
				User: &bot.User{
					Id:     content.Data,
					Name:   u.Name,
					Avatar: u.Avatar,
				},
			})
		}
	}
	return &result, nil
}

func (d *DummyAdapter) SendPrivateMessage(userId string, msg *bot.ClientMessage) (string, error) {
	// d.wrapper.queryPrivateSession(d.wrapper.Bot, userId)
	// d.wrapper.pushMessage()
	puts := covertMessage(msg.Message)

	return d.wrapper.SendPrivateMessage(userId, msg.QuoteID, puts)
}

func (d *DummyAdapter) SendGroupMessage(groupID string, msg *bot.ClientMessage) (string, error) {
	puts := covertMessage(msg.Message)

	return d.wrapper.SendGroupMessage(groupID, msg.QuoteID, puts)
}

func covertMessage(msg *bot.Contents) []*bot.RawContent {
	puts := make([]*bot.RawContent, 0)
	for _, content := range *msg {
		switch it := content.(type) {
		case bot.ContentText:
			puts = append(puts, &bot.RawContent{
				Type: "text", Data: it.Text,
			})
		case bot.ContentAt:
			puts = append(puts, &bot.RawContent{
				Type: "at", Data: it.Uid,
			})
		case bot.ContentImage:
			puts = append(puts, &bot.RawContent{
				Type: "image", Data: it.URL,
			})
		}
	}
	return puts
}

func NewDataUriContentImage(dataURI string) (*bot.ContentImage, error) {
	if !strings.HasPrefix(dataURI, "data:") {
		return nil, errors.New("invalid data URI format: missing 'data:' prefix")
	}
	parts := strings.SplitN(dataURI[5:], ",", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid data URI format: missing comma separator")
	}

	mediaTypePart := parts[0]
	dataPart := parts[1]
	mediaType := strings.SplitN(mediaTypePart, ";", 2)
	if len(mediaType) == 2 && mediaType[1] == "base64" {
		image := bot.NewBase64ContentImage(
			url.QueryEscape(mediaType[0]),
			dataPart,
			"",
		)
		return &image, nil
	}
	return nil, errors.New("unsupported data URI format: must be base64 encoded")
}
