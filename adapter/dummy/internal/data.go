package internal

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/GreekMilkBot/GreekMilkBot/adapter/dummy/internal/server"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"github.com/gorilla/websocket"
)

type Tree struct {
	*http.ServeMux `json:"-"`
	router         map[string]chan string
	mutex          *sync.RWMutex
	broker         *Broker[string]

	Self string `json:"self"`
	Bot  string `json:"bot"`

	Server         *server.Server                     `json:"server"`
	BindBotMessage func(resp server.QueryMessageResp) `json:"-"`
}

func NewTree() *Tree {
	t := &Tree{
		ServeMux: http.NewServeMux(),
		router:   make(map[string]chan string),
		mutex:    new(sync.RWMutex),
		Server:   server.NewServer(),
		broker:   NewBroker[string](),
		Self:     "",
		Bot:      "",
	}
	t.HandleFunc("/api/ws", t.handleEvent)
	t.HandleFunc("/api/self", handleJSON(t.handleSelf))
	t.HandleFunc("/api/sessions", handleJSON(t.handleSessions))
	t.HandleFunc("/api/user", handleJSON(t.handleUser))
	t.HandleFunc("/api/group", handleJSON(t.handleGroup))
	t.HandleFunc("/api/messages", handleJSON(t.handleMessages))
	t.HandleFunc("/api/message", handleJSON(t.handleMessage))
	t.HandleFunc("/api/send", handleJSON(t.handleSend))
	go t.broker.Start()
	return t
}

func (t *Tree) Close() error {
	t.broker.Stop()
	return nil
}

func (t *Tree) handleSelf(_ *http.Request) (any, error) {
	return t.Server.GetUser(t.Self)
}

func (t *Tree) handleSessions(_ *http.Request) (any, error) {
	return t.Server.GetSessions(t.Self), nil
}

func (t *Tree) handleMessages(r *http.Request) (any, error) {
	sid := r.URL.Query().Get("id")
	return t.Server.GetMessages(t.Self, sid)
}

func (t *Tree) handleUser(r *http.Request) (any, error) {
	uid := r.URL.Query().Get("id")
	return t.Server.GetUser(uid)
}

func (t *Tree) handleGroup(r *http.Request) (any, error) {
	mid := r.URL.Query().Get("id")
	return t.Server.GetGuild(mid)
}

func (t *Tree) handleMessage(r *http.Request) (any, error) {
	gid := r.URL.Query().Get("id")
	return t.Server.QueryMessage(gid)
}

func (t *Tree) SendPrivateMessage(userID string, content []*bot.RawContent) (string, error) {
	id, err := t.Server.GetOrCreatePrivateSessionID(t.Bot, userID)
	if err != nil {
		return "", err
	}
	return t.Server.AddMessage(server.AddMessageReq{
		UserID:    t.Bot,
		SessionID: id,
		ReferID:   "",
		Content:   content,
	})
}

func (t *Tree) SendGroupMessage(id string, content []*bot.RawContent) (string, error) {
	sid, err := t.Server.GetSessionIDGroupByID(id)
	if err != nil {
		return "", err
	}
	return t.Server.AddMessage(server.AddMessageReq{
		UserID:    t.Bot,
		SessionID: sid,
		ReferID:   "",
		Content:   content,
	})
}

func (t *Tree) handleSend(r *http.Request) (any, error) {
	req := server.AddMessageReq{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	req.UserID = t.Self
	id, err := t.Server.AddMessage(req)
	message, err := t.Server.QueryMessage(id)
	if err != nil {
		return nil, err
	}
	botMessage := t.BindBotMessage
	if botMessage != nil {
		go botMessage(*message)
	}
	t.broker.Publish("{}")
	return nil, err
}

var wsCfg = websocket.Upgrader{}

func (t *Tree) handleEvent(w http.ResponseWriter, r *http.Request) {
	c, err := wsCfg.Upgrade(w, r, nil)
	if err != nil {
		log.Warnf("upgrade error: %s", err)
		return
	}
	defer c.Close()
	subscribe := t.broker.Subscribe()
	defer t.broker.Unsubscribe(subscribe)
	for {
		select {
		case message := <-subscribe:
			err = c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Warnf("write error: %s", err)
				break
			}
		}
	}
}

func handleJSON(eval func(r *http.Request) (any, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		a, err := eval(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(a)
	}
}
