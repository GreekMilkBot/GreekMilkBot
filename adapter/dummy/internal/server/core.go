package server

import (
	"errors"
	"github.com/GreekMilkBot/GreekMilkBot/bot"
	"time"
)

type Server struct {
	Users    map[string]*User    `json:"users"`
	Guilds   map[string]*Guild   `json:"guilds"`
	Sessions map[string]*Session `json:"sessions"`
}

func NewServer() *Server {
	return &Server{
		Users:    make(map[string]*User),
		Guilds:   make(map[string]*Guild),
		Sessions: make(map[string]*Session),
	}
}

type User struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type Guild struct {
	Name   string                `json:"name"`
	Avatar string                `json:"avatar"`
	Users  map[string]*GroupUser `json:"users"` // users
}

type GroupUser struct {
	Name string `json:"name"`
}

type Message struct {
	ID       string          `json:"id"`
	Sender   string          `json:"sender"`
	CreateAt time.Time       `json:"created"`
	Content  *MessageContent `json:"content"`
}
type MessageContent struct {
	Refer   string            `json:"refer,omitempty"`
	Message []*bot.RawContent `json:"message"`
}

type Session struct {
	Object   []string   `json:"child"` // 如果是私聊则为 2 个用户 ，否则为群聊地址
	Messages []*Message `json:"messages"`
}

func (s *Session) IsGroup() bool {
	return len(s.Object) == 1
}
func (s *Session) GroupID() string {
	return s.Object[0]
}
func (s *Session) TargetUserID(self string) string {
	if s.IsGroup() {
		return ""
	}
	for _, i := range s.Object {
		if i != self {
			return i
		}
	}
	return ""
}

type SessionResp struct {
	Name        string       `json:"name"`         // 会话名称
	SType       string       `json:"type"`         // 会话类型
	Target      string       `json:"target"`       // 对象 （目标用户/群组）
	LastMessage *MessageResp `json:"last_message"` // 最后一条消息
}

// GetSessions 查询当前用户可用的会话
func (s *Server) GetSessions(userID string) map[string]SessionResp {
	result := make(map[string]SessionResp)
	for id, session := range s.Sessions {
		resp := SessionResp{}
		isGroup := session.IsGroup()
		groupUsers := make(map[string]*GroupUser)
		if isGroup {
			groupID := session.GroupID()
			guild := s.Guilds[groupID]
			groupUsers = guild.Users
			if guild.Users[userID] == nil {
				continue
			}
			resp.SType = "group"
			resp.Target = groupID
			resp.Name = guild.Name
		} else {
			userID := session.TargetUserID(userID)
			if userID == "" {
				continue
			}
			resp.SType = "userID"
			resp.Target = userID
			user, err := s.GetUser(id)
			if err != nil {
				panic(err)
			}
			resp.Name = user.Name
		}
		lastMsg := session.Messages[len(session.Messages)-1]
		lastMsgUser, err := s.GetUser(lastMsg.Sender)
		if err != nil {
			panic(err)
		}
		var aliasName string
		if isGroup {
			aliasName = groupUsers[lastMsgUser.ID].Name
		}
		resp.LastMessage = &MessageResp{
			ID: lastMsg.ID,
			Sender: UserResp{
				Name:      lastMsgUser.Name,
				Avatar:    lastMsgUser.Avatar,
				AliasName: aliasName,
			},
			ReferID:  lastMsg.Content.Refer,
			CreateAt: lastMsg.CreateAt,
			Content:  lastMsg.Content.Message,
		}
		result[id] = resp

	}
	return result
}

type UserResp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AliasName string `json:"alias,omitempty"`
	Avatar    string `json:"avatar"`
}

func (s *Server) GetUser(userID string) (UserResp, error) {
	u := s.Users[userID]
	if u == nil {
		return UserResp{}, errors.New("userID not found")
	}
	return UserResp{
		ID:     userID,
		Name:   u.Name,
		Avatar: u.Avatar,
	}, nil
}

type MessageResp struct {
	ID       string            `json:"id"`
	Sender   UserResp          `json:"sender"`
	ReferID  string            `json:"refer_id"`
	Content  []*bot.RawContent `json:"content"`
	CreateAt time.Time         `json:"created"`
}

func (s *Server) GetMessages(userID, sessionID string) ([]MessageResp, error) {
	result := make([]MessageResp, 0)
	session := s.Sessions[sessionID]
	if session == nil {
		return nil, errors.New("session not found")
	}
	isGroup := session.IsGroup()
	var groupUser *Guild
	if isGroup {
		groupUser = s.Guilds[session.GroupID()]
		if groupUser.Users[userID] == nil {
			return nil, errors.New("userID not allowed group session")
		}
	} else {
		if session.TargetUserID(userID) == "" {
			return nil, errors.New("userID not allowed private session")
		}
	}

	for _, message := range session.Messages {
		msgUser, err := s.GetUser(message.Sender)
		if err != nil {
			return nil, err
		}
		if isGroup && groupUser != nil {
			msgUser.AliasName = groupUser.Users[message.Sender].Name
		}
		result = append(result, MessageResp{
			ID:       message.ID,
			Sender:   msgUser,
			ReferID:  message.Content.Refer,
			CreateAt: message.CreateAt,
			Content:  message.Content.Message,
		})
	}
	return result, nil
}
