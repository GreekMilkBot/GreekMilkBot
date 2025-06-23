package dummy

import (
	"encoding/json"
	"fmt"
	"github.com/GreekMilkBot/GreekMilkBot/adapter/dummy/static"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"strings"
	"sync"
	"sync/atomic"

	"net/http"
	"time"
)

type CustomTime time.Time

const ctLayout = "2006-01-02 15:04:05"

// UnmarshalJSON Parses the json string in the custom format
func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(ctLayout, s)
	*ct = CustomTime(nt)
	return
}

// MarshalJSON writes a quoted string in the custom format
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(ct.String()), nil
}

// String returns the time in the custom format
func (ct *CustomTime) String() string {
	t := time.Time(*ct)
	return fmt.Sprintf("%q", t.Format(ctLayout))
}

type Tree struct {
	*http.ServeMux `json:"-"`
	router         map[string]chan string
	mutex          *sync.RWMutex

	Self string `json:"self"`
	Bot  string `json:"bot"`

	Users    map[string]*User    `json:"users"`
	Guilds   map[string]*Guild   `json:"guilds"`
	Sessions map[string]*Session `json:"sessions"`

	randomID *atomic.Int64
}

func NewTree() *Tree {
	t := &Tree{
		ServeMux: http.NewServeMux(),
		router:   make(map[string]chan string),
		mutex:    new(sync.RWMutex),
		Users:    make(map[string]*User),
		Guilds:   make(map[string]*Guild),
		Sessions: make(map[string]*Session),

		randomID: &atomic.Int64{},
		Self:     "",
		Bot:      "",
	}
	t.HandleFunc("/api/self", t.handleSelf)
	t.HandleFunc("/api/users", t.handleUsers)
	t.HandleFunc("/api/groups", t.handleGroups)
	t.HandleFunc("/api/sessions", t.handleSessions)
	t.HandleFunc("/api/messages", t.handleMessages)
	t.HandleFunc("/api/message", t.handleMessage)
	t.HandleFunc("/api/send", t.handleSend)
	t.Handle("/", http.FileServerFS(static.FS))
	return t
}

func withID[T any](data map[string]T) (map[string]map[string]any, error) {
	result := make(map[string]map[string]any)
	for id, datum := range data {
		marshal, err := json.Marshal(datum)
		item := make(map[string]any)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(marshal, &item)
		if err != nil {
			return nil, err
		}
		item["id"] = id
		result[id] = item
	}
	return result, nil
}

type Session struct {
	SType    string     `json:"type"` // private or group
	Target   string     `json:"target"`
	Messages []*Message `json:"messages"`
}

type User struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type Guild struct {
	bot.Guild `json:",inline"`
	Users     map[string]*GroupUser `json:"users"` // user id
}

type GroupUser struct {
	Name string `json:"name"`
}

type Message struct {
	ID       string          `json:"id"`
	Sender   string          `json:"sender"`
	CreateAt CustomTime      `json:"created"`
	Content  *MessageContent `json:"content"`
}
type MessageContent struct {
	Refer   string            `json:"refer,omitempty"`
	Message []*bot.RawContent `json:"message"`
}

type RequestMsg struct {
	Session         string `json:"session"`
	Sender          string `json:"user"`
	*MessageContent `json:",inline"`
}

func (t *Tree) handleUsers(writer http.ResponseWriter, request *http.Request) {
	data, err := withID(t.Users)
	if err != nil {
		log.Errorf("handleUsers: %v", err)
	}

	if err := json.NewEncoder(writer).Encode(data); err != nil {
		log.Debugf("websocket encode error: %v", err)
	}
}

func (t *Tree) handleSessions(writer http.ResponseWriter, request *http.Request) {
	result := make(map[string]map[string]any)
	for s, session := range t.Sessions {
		if session.SType == "group" && t.Guilds[session.Target].Users[t.Self] == nil {
			continue
		}
		m := make(map[string]any)
		m["id"] = s
		m["type"] = session.SType
		m["target"] = session.Target
		messages := t.Sessions[s].Messages
		m["lastMessage"] = messages[len(messages)-1]
		result[s] = m
	}
	_ = json.NewEncoder(writer).Encode(result)
}
func (t *Tree) String() string {
	return fmt.Sprintf("包含 %d 位用户", len(t.Users))
}

func (t *Tree) handleGroups(writer http.ResponseWriter, request *http.Request) {
	result := make(map[string]*Guild)
	for s, guild := range t.Guilds {
		if guild.Users[t.Self] != nil {
			result[s] = guild
		}
	}
	id, err := withID(result)
	if err != nil {
		log.Errorf("handleGroups: %v", err)
		return
	}
	_ = json.NewEncoder(writer).Encode(id)
}

func (t *Tree) handleSelf(writer http.ResponseWriter, request *http.Request) {
	user := t.Users[t.Self]
	_ = json.NewEncoder(writer).Encode(bot.User{
		Id:     t.Self,
		Name:   user.Name,
		Avatar: user.Avatar,
	})
}

func (t *Tree) handleMessages(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	sid := query.Get("sid")
	message := t.Sessions[sid].Messages
	if message == nil {
		http.NotFound(writer, request)
	}
	result := make([]*Message, 0, len(message))
	for _, m := range message {
		result = append(result, &Message{
			ID:       fmt.Sprintf("%s@%s", sid, m.ID),
			Sender:   m.Sender,
			CreateAt: m.CreateAt,
			Content:  m.Content,
		})
	}
	_ = json.NewEncoder(writer).Encode(result)
}
func (t *Tree) handleMessage(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	id := query.Get("id")
	sid, mid, found := strings.Cut(id, "@")
	if !found {
		http.NotFound(writer, request)
		return
	}
	for _, message := range t.Sessions[sid].Messages {
		if message.ID == mid {
			_ = json.NewEncoder(writer).Encode(message)
			return
		}
	}
	http.NotFound(writer, request)
	return
}

func (t *Tree) handleSend(writer http.ResponseWriter, request *http.Request) {
	ref := &RequestMsg{}
	if err := json.NewDecoder(request.Body).Decode(ref); err != nil {
		log.Errorf("handleSend: %v", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	t.Sessions[ref.Session].Messages = append(t.Sessions[ref.Session].Messages, &Message{
		ID:       fmt.Sprintf("%s@%d", ref.Session, t.randomID.Add(1)),
		Sender:   t.Self,
		CreateAt: CustomTime(time.Now()),
		Content:  ref.MessageContent,
	})
}
