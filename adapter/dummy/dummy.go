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
		tree := internal.NewTree()
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
			if err := json.Unmarshal(data, tree); err != nil {
				return nil, err
			}

		}
		tree.Handle("/", http.FileServerFS(static.FS))
		go func() {
			select {
			case <-ctx.Done():
				_ = listen.Close()
			}
		}()
		go func() {
			_ = http.Serve(listen, tree)
		}()
		return &DummyAdapter{
			tree,
		}, nil
	})
}

type DummyAdapter struct {
	tree *internal.Tree
}

func (d *DummyAdapter) Bind(ctx *bot.Bus) error {
	d.tree.BindBotMessage = func(msg server.QueryMessageResp) {
		message, err := d.Dummy2Message(msg, 5)
		if err != nil {
			log.Warnf("message covert error %s", err.Error())
			return
		}
		ctx.SendMessage(*message)
	}
	ctx.CallFunc("send_private_msg", d.sendPrivateMessage)
	ctx.CallFunc("send_group_msg", d.sendGroupMessage)
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
		Owner: &bot.User{
			Id:     content.Sender.ID,
			Name:   content.Sender.Name,
			Avatar: content.Sender.Avatar,
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
	}
	if content.ReferID != "" {
		query, err := d.tree.Server.QueryMessage(content.ReferID)
		if err != nil {
			return nil, err
		}
		botMsg, err := d.Dummy2Message(*query, depth)
		if err != nil {
			return nil, err
		}
		result.Quote = botMsg
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
			u, err := d.tree.Server.GetUser(content.Data)
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

func (d *DummyAdapter) sendPrivateMessage(userId string, msg *bot.Contents) (string, error) {
	// d.tree.queryPrivateSession(d.tree.Bot, userId)
	// d.tree.pushMessage()
	puts := covertMessage(msg)

	return d.tree.SendPrivateMessage(userId, puts)
}

func (d *DummyAdapter) sendGroupMessage(groupID string, msg *bot.Contents) (string, error) {
	puts := covertMessage(msg)

	return d.tree.SendGroupMessage(groupID, puts)
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
