package internal

import (
	"encoding/json"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/dummy/internal/server"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type Tree struct {
	*http.ServeMux `json:"-"`
	router         map[string]chan string
	mutex          *sync.RWMutex
	broker         *Broker[string]

	Self string `json:"self"`
	Bot  string `json:"bot"`

	Server *server.Server `json:"server"`

	BindBotMessage func(resp server.MessageResp) `json:"-"`
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
	t.HandleFunc("/api/messages", handleJSON(t.handleMessages))
	//t.HandleFunc("/api/message", t.handleMessage)
	//t.HandleFunc("/api/send", t.handleSend)
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
	sid := r.URL.Query().Get("sid")
	return t.Server.GetMessages(t.Self, sid)
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
		}
		_ = json.NewEncoder(w).Encode(a)
	}
}
