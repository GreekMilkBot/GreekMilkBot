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

	Users    map[string]*User      `json:"users"`
	Guilds   map[string]*Guild     `json:"guilds"`
	Sessions map[string]*Session   `json:"sessions"`
	Messages map[string][]*Message `json:"messages"`

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
		Messages: make(map[string][]*Message),

		randomID: &atomic.Int64{},
		Self:     "",
	}
	t.HandleFunc("/api/self", t.handleSelf)
	t.HandleFunc("/api/users", t.handleUsers)
	t.HandleFunc("/api/groups", t.handleGroups)
	t.HandleFunc("/api/sessions", t.handleSessions)
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
	SType  string `json:"type"` // private or group
	Target string `json:"target"`
}

type User struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type Guild struct {
	bot.Guild `json:",inline"`
	Users     []*GroupUser `json:"users"` // user id
}

type GroupUser struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type Message struct {
	ID       string          `json:"id"`
	Sender   string          `json:"sender"`
	CreateAt CustomTime      `json:"created"`
	Content  *MessageContent `json:"content"`
}
type MessageContent struct {
	Refer   *Refer            `json:"refer,omitempty"`
	Message []*bot.RawContent `json:"message"`
}
type Refer struct {
	Sid string `json:"sid,omitempty"`
	Mid string `json:"mid,omitempty"`
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
		m := make(map[string]any)
		m["id"] = s
		m["type"] = session.SType
		m["target"] = session.Target
		messages := t.Messages[s]
		m["lastMessage"] = messages[len(messages)-1]
		result[s] = m
	}
	_ = json.NewEncoder(writer).Encode(result)
}
func (t *Tree) String() string {
	return fmt.Sprintf("包含 %d 位用户", len(t.Users))
}

func (t *Tree) handleGroups(writer http.ResponseWriter, request *http.Request) {
	id, err := withID(t.Guilds)
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

func (t *Tree) handleMessage(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	query.Get("sid")
	message := t.Messages[query.Get("sid")]
	if message == nil {
		http.NotFound(writer, request)
	}
	_ = json.NewEncoder(writer).Encode(message)
}

func (t *Tree) handleSend(writer http.ResponseWriter, request *http.Request) {
	ref := &RequestMsg{}
	if err := json.NewDecoder(request.Body).Decode(ref); err != nil {
		log.Errorf("handleSend: %v", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	t.Messages[ref.Session] = append(t.Messages[ref.Session], &Message{
		ID:       fmt.Sprintf("%d", t.randomID.Add(1)),
		Sender:   t.Self,
		CreateAt: CustomTime(time.Now()),
		Content:  ref.MessageContent,
	})
}
