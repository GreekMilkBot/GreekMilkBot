package dummy

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GreekMilkBot/GreekMilkBot/pkg/tools"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models"

	"github.com/GreekMilkBot/GreekMilkBot/adapters/dummy/internal"
	"github.com/GreekMilkBot/GreekMilkBot/adapters/dummy/internal/server"
	"github.com/GreekMilkBot/GreekMilkBot/adapters/dummy/static"
	"github.com/GreekMilkBot/GreekMilkBot/pkg"
	"github.com/GreekMilkBot/GreekMilkBot/pkg/log"
)

//go:embed default.json
var defCfg []byte

func init() {
	core.RegisterAdapter("dummy", func(ctx context.Context, url url.URL) (core.Adapter, error) {
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
			wrapper: wrapper,
		}, nil
	})
}

type DummyAdapter struct {
	wrapper        *internal.Wrapper
	imageFormatter core.ResourceFormatter
	ctx            *core.AdapterBus
}

func (d *DummyAdapter) Metadata(scheme, body string) (*models.Metadata, error) {
	if scheme != "image" {
		return nil, errors.New("unknown scheme :" + scheme)
	}
	parse, err := url.Parse(body)
	if err != nil {
		return nil, err
	}
	switch parse.Scheme {
	case "base64":
		data, err := base64.URLEncoding.DecodeString(parse.Path)
		if err != nil {
			return nil, err
		}
		return &models.Metadata{
			Name:      "",
			Size:      int64(len(data)),
			MediaType: parse.Query().Get("media-type"),
		}, nil
	case "http", "https":
		return models.HttpMetadata(http.DefaultClient, body)
	default:
		return nil, errors.New("unknown scheme :" + parse.Scheme)
	}
}

func (d *DummyAdapter) Reader(scheme, body string) (io.ReadCloser, error) {
	if scheme != "image" {
		return nil, errors.New("unknown scheme :" + scheme)
	}
	parse, err := url.Parse(body)
	if err != nil {
		return nil, err
	}
	switch parse.Scheme {
	case "base64":
		data, err := base64.URLEncoding.DecodeString(parse.Path)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewBuffer(data)), nil
	case "http", "https":
		return models.HttpReader(http.DefaultClient, body)
	default:
		return nil, errors.New("unknown scheme :" + parse.Scheme)
	}
}

func (d *DummyAdapter) Bind(ctx *core.AdapterBus) error {
	d.ctx = ctx
	d.imageFormatter = ctx.BindResource("image", d)
	d.wrapper.BindBotMessage = func(msg server.QueryMessageResp) {
		message, err := d.Dummy2Message(msg, 5)
		if err != nil {
			log.Warnf("message covert error %s", err.Error())
			return
		}
		ctx.SendMessage(*message)
	}
	ctx.BindTools(tools.Sender(d))
	return nil
}

func (d *DummyAdapter) Dummy2Message(msg server.QueryMessageResp, depth int) (*models.Message, error) {
	if depth == 0 {
		return nil, errors.New("depth is zero")
	}
	content := msg.Content
	depth = depth - 1
	result := models.Message{
		ID: content.ID,
		Owner: &models.GuildMember{
			User: &models.User{
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
		result.Guild = &models.Guild{
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
			result.Content = append(result.Content, models.ContentText{Text: content.Data})
		case "image":
			bind, err := d.imageFormatter(content.Data)
			if err != nil {
				return nil, err
			}
			result.Content = append(result.Content, models.ContentImage{
				Resource: bind,
				Summary:  "",
			})
		case "at":
			u, err := d.wrapper.Server.GetUser(content.Data)
			if err != nil {
				return nil, err
			}
			result.Content = append(result.Content, models.ContentAt{
				Uid: content.Data,
				User: &models.User{
					Id:     content.Data,
					Name:   u.Name,
					Avatar: u.Avatar,
				},
			})
		}
	}
	return &result, nil
}

func (d *DummyAdapter) SendPrivateMessage(userId string, msg *tools.SenderMessage) (string, error) {
	// d.wrapper.queryPrivateSession(d.wrapper.Bot, userId)
	// d.wrapper.pushMessage()
	puts := d.covertMessage(msg.Message)

	return d.wrapper.SendPrivateMessage(userId, msg.QuoteID, puts)
}

func (d *DummyAdapter) SendGroupMessage(groupID string, msg *tools.SenderMessage) (string, error) {
	puts := d.covertMessage(msg.Message)
	return d.wrapper.SendGroupMessage(groupID, msg.QuoteID, puts)
}

func (d *DummyAdapter) covertMessage(msg *models.Contents) []*models.RawContent {
	puts := make([]*models.RawContent, 0)
	for _, content := range *msg {
		switch it := content.(type) {
		case models.ContentText:
			puts = append(puts, &models.RawContent{
				Type: "text", Data: it.Text,
			})
		case models.ContentAt:
			puts = append(puts, &models.RawContent{
				Type: "at", Data: it.Uid,
			})
		case models.ContentImage:
			meta, err := d.ctx.ResourceMeta(&it.Resource)
			if err != nil {
				continue
			}
			blob, err := d.ctx.ResourceBlob(&it.Resource)
			if err != nil {
				continue
			}
			all, err := io.ReadAll(blob)
			blob.Close()
			if err != nil {
				continue
			}
			puts = append(puts, &models.RawContent{
				Type: "image", Data: fmt.Sprintf("data:%s;base64,%s", meta.MediaType, base64.URLEncoding.EncodeToString(all)),
			})
		}
	}
	return puts
}
