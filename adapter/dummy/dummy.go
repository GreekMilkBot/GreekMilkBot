package dummy

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/dummy/internal"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/dummy/static"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/gmb"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	//d.tree.BindBotMessage = func(msg internal.ResponseMsg) {
	//	message, err := d.Dummy2Message(msg, 5)
	//	if err != nil {
	//		log.Warnf("message covert error %s", err.Error())
	//		return
	//	}
	//	ctx.SendMessage(*message)
	//}
	//ctx.CallFunc("send_private_msg", d.sendPrivateMessage)
	//ctx.CallFunc("send_group_msg", d.sendGroupMessage)
	return nil
}

//func (d *DummyAdapter) Dummy2Message(msg internal.ResponseMsg, depth int) (*bot.Message, error) {
//	if depth == 0 {
//		return nil, errors.New("depth is zero")
//	}
//	depth = depth - 1
//	user := d.tree.Users[msg.Sender]
//	result := bot.Message{
//		ID: msg.Sender,
//		Owner: &bot.User{
//			Id:     msg.Sender,
//			Name:   user.Name,
//			Avatar: user.Avatar,
//		},
//		Created: msg.Created,
//		Updated: msg.Created,
//	}
//	if session := d.tree.Sessions[msg.Session]; session.SType == "group" {
//		guild := d.tree.Guilds[session.Target]
//		result.Guild = &bot.Guild{
//			Id:     session.Target,
//			Name:   guild.Name,
//			Avatar: guild.Avatar,
//		}
//	}
//	if msg.Refer != "" {
//		sid, message := d.tree.GetMessage(msg.Refer)
//		botMsg, err := d.Dummy2Message(internal.ResponseMsg{
//			RequestMsg: internal.RequestMsg{
//				Session:        sid,
//				Sender:         message.Sender,
//				MessageContent: message.Content,
//			},
//			ID:      message.ID,
//			Created: time.Time(message.CreateAt),
//		}, depth)
//		if err != nil {
//			return nil, err
//		}
//		result.Quote = botMsg
//	}
//	for _, content := range msg.MessageContent.Message {
//		switch content.Type {
//		case "text":
//			result.Content = append(result.Content, bot.ContentText{Text: content.Data})
//		case "image":
//			if strings.HasPrefix(content.Data, "data:") {
//				image, err := NewDataUriContentImage(content.Data)
//				if err != nil {
//					return nil, err
//				}
//				result.Content = append(result.Content, image)
//			} else {
//				result.Content = append(result.Content, bot.ContentImage{
//					URL:     content.Data,
//					Summary: "",
//				})
//			}
//		case "at":
//			u := d.tree.Users[content.Data]
//			result.Content = append(result.Content, bot.ContentAt{
//				Uid: content.Data,
//				User: &bot.User{
//					Id:     content.Data,
//					Name:   u.Name,
//					Avatar: u.Avatar,
//				},
//			})
//		}
//	}
//	return &result, nil
//}
//
//func (d *DummyAdapter) sendPrivateMessage(userId string, msg *bot.Contents) (string, error) {
//	//d.tree.queryPrivateSession(d.tree.Bot, userId)
//	//d.tree.pushMessage()
//	return a.sendMessage(userId, "", msg)
//}
//
//func (d *DummyAdapter) sendGroupMessage(groupID string, msg *bot.Contents) (string, error) {
//	return a.sendMessage("", groupID, msg)
//}

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
