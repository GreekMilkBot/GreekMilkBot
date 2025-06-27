package server

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/GreekMilkBot/GreekMilkBot/pkg/models/bot"
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
	CreateAt *CustomTime     `json:"created"`
	Content  *MessageContent `json:"content"`
}
type MessageRefer struct {
	SessionID string `json:"sid"`
	MessageID string `json:"mid"`
}
type MessageContent struct {
	Refer   *MessageRefer     `json:"refer,omitempty"`
	Message []*bot.RawContent `json:"message"`
}

func (receiver *MessageRefer) toReferID() string {
	if receiver == nil {
		return ""
	}
	return MessageID(receiver.SessionID, receiver.MessageID)
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
	if !slices.Contains(s.Object, self) {
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
	ID          string       `json:"id"`
	Name        string       `json:"name"`         // 会话名称
	SType       string       `json:"type"`         // 会话类型
	Avatar      string       `json:"avatar"`       // 头像
	Target      string       `json:"target"`       // 对象 （目标用户/群组）
	LastMessage *MessageResp `json:"last_message"` // 最后一条消息
}

// GetSessions 查询当前用户可用的会话
func (s *Server) GetSessions(userID string) map[string]*SessionResp {
	result := make(map[string]*SessionResp)
	for sessionID, session := range s.Sessions {
		resp := SessionResp{
			ID: sessionID,
		}
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
			resp.Avatar = guild.Avatar
		} else {
			targetUserID := session.TargetUserID(userID)
			if targetUserID == "" {
				continue
			}
			resp.SType = "private"
			resp.Target = targetUserID
			user, err := s.GetUser(targetUserID)
			if err != nil {
				panic(err)
			}
			resp.Name = user.Name
			resp.Avatar = user.Avatar
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
			ID: MessageID(sessionID, lastMsg.ID),
			Sender: UserResp{
				ID:        lastMsgUser.ID,
				Name:      lastMsgUser.Name,
				Avatar:    lastMsgUser.Avatar,
				AliasName: aliasName,
			},
			ReferID:  lastMsg.Content.Refer.toReferID(),
			CreateAt: time.Time(*lastMsg.CreateAt),
			Content:  lastMsg.Content.Message,
		}
		result[sessionID] = &resp

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
		return UserResp{}, errors.New("userID not found " + userID)
	}
	return UserResp{
		ID:     userID,
		Name:   u.Name,
		Avatar: u.Avatar,
	}, nil
}

type GuildResp struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	Avatar string               `json:"avatar"`
	Users  map[string]*UserResp `json:"users"`

	MessageCount int `json:"message_count"`
}

func (s *Server) GetGuild(gid string) (*GuildResp, error) {
	guild := s.Guilds[gid]
	if guild == nil {
		return nil, errors.New("guild not found " + gid)
	}
	resp := GuildResp{
		ID:     gid,
		Name:   guild.Name,
		Avatar: guild.Avatar,
		Users:  make(map[string]*UserResp),
	}
	for id, user := range guild.Users {
		gUser, err := s.GetUser(id)
		if err != nil {
			panic(err)
		}
		gUser.AliasName = user.Name
		resp.Users[id] = &gUser
		for _, session := range s.Sessions {
			if session.IsGroup() && session.GroupID() == gid {
				resp.MessageCount = len(session.Messages)
			}
		}
	}
	return &resp, nil
}

type MessagesResp struct {
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	Count   int           `json:"count"`
	Content []MessageResp `json:"content"`
}
type MessageResp struct {
	ID       string            `json:"id"`
	Sender   UserResp          `json:"sender"`
	ReferID  string            `json:"refer_id"`
	Content  []*bot.RawContent `json:"content"`
	CreateAt time.Time         `json:"created"`
}

func (s *Server) GetMessages(userID, sessionID string) (*MessagesResp, error) {
	result := &MessagesResp{
		Content: make([]MessageResp, 0),
	}
	session := s.Sessions[sessionID]
	if session == nil {
		return nil, errors.New("session not found")
	}
	result.Count = len(session.Messages)
	isGroup := session.IsGroup()
	var guild *Guild
	if isGroup {
		guild = s.Guilds[session.GroupID()]
		if guild.Users[userID] == nil {
			return nil, errors.New("userID not allowed group session")
		}
		result.Name = guild.Name
		result.Type = "group"
	} else {
		targetUserID := session.TargetUserID(userID)
		if targetUserID == "" {
			return nil, errors.New("userID not allowed private session")
		}
		user, err := s.GetUser(targetUserID)
		if err != nil {
			panic(err)
		}
		result.Name = user.Name
		result.Type = "private"
	}

	for _, message := range session.Messages {
		msgUser, err := s.GetUser(message.Sender)
		if err != nil {
			return nil, err
		}
		if isGroup && guild != nil {
			msgUser.AliasName = guild.Users[message.Sender].Name
		}
		result.Content = append(result.Content, MessageResp{
			ID:       MessageID(sessionID, message.ID),
			Sender:   msgUser,
			ReferID:  message.Content.Refer.toReferID(),
			CreateAt: time.Time(*message.CreateAt),
			Content:  message.Content.Message,
		})
	}
	return result, nil
}

type QueryMessageResp struct {
	Type         string       `json:"type"`
	TargetID     string       `json:"target"`
	TargetName   string       `json:"name"`
	TargetAvatar string       `json:"avatar"`
	Content      *MessageResp `json:"content"`
}

func (s *Server) QueryMessage(mid string) (*QueryMessageResp, error) {
	sid, mid, ok := ParseMessageID(mid)
	if !ok {
		return nil, errors.New("invalid message id")
	}
	session := s.Sessions[sid]
	if session == nil {
		return nil, errors.New("session not found")
	}
	var message *Message
	for _, m := range session.Messages {
		if m.ID == mid {
			message = m
			break
		}
	}
	if message == nil {
		return nil, errors.New("message not found")
	}
	result := &QueryMessageResp{}
	msgUser, err := s.GetUser(message.Sender)
	if err != nil {
		return nil, err
	}
	if session.IsGroup() {
		result.TargetID = session.GroupID()
		result.Type = "group"
		guild := s.Guilds[session.GroupID()]
		result.TargetAvatar = guild.Avatar
		msgUser.AliasName = guild.Users[message.Sender].Name
		result.TargetName = guild.Name
	} else {
		user := s.Users[message.Sender]
		result.TargetName = user.Name
		result.TargetAvatar = user.Avatar
		result.TargetID = message.Sender
		result.Type = "private"
	}
	result.Content = &MessageResp{
		ID:       MessageID(sid, message.ID),
		Sender:   msgUser,
		ReferID:  message.Content.Refer.toReferID(),
		CreateAt: time.Time(*message.CreateAt),
		Content:  message.Content.Message,
	}
	return result, nil
}

type AddMessageReq struct {
	UserID    string            `json:"user_id"`
	SessionID string            `json:"session_id"`
	ReferID   string            `json:"refer_id"`
	Content   []*bot.RawContent `json:"content"`
}

func (s *Server) AddMessage(msg AddMessageReq) (string, error) {
	user := s.Users[msg.UserID]
	if user == nil {
		return "", errors.New("user not found")
	}
	session := s.Sessions[msg.SessionID]
	if session == nil {
		return "", errors.New("session not found")
	}
	var guild *Guild
	if session.IsGroup() {
		guild = s.Guilds[session.GroupID()]
		if guild == nil {
			panic("guild not found")
		}
		groupUser := guild.Users[msg.UserID]
		if groupUser == nil {
			return "", errors.New("user not allowed group")
		}
	}
	if msg.ReferID != "" {
		_, err := s.QueryMessage(msg.ReferID)
		if err != nil {
			return "", err
		}
	}
	if msg.Content == nil || len(msg.Content) == 0 {
		return "", errors.New("message content is empty")
	}
	for _, content := range msg.Content {
		switch content.Type {
		case "text":
		case "image":
		case "at":
			if guild == nil {
				return "", errors.New("private session not allowed @")
			}
			if guild.Users[content.Data] == nil {
				return "", errors.New("@ user not found: " + content.Data)
			}
			if content.Data == "" {
				return "", errors.New("message content at is empty")
			}
		default:
			return "", errors.New("invalid content type:" + content.Type)
		}
	}
	customTime := CustomTime(time.Now())
	next := &Message{
		ID:       fmt.Sprintf("10%d", len(session.Messages)),
		Sender:   msg.UserID,
		CreateAt: &customTime,
		Content: &MessageContent{
			Message: msg.Content,
		},
	}
	if msg.ReferID != "" {
		sid, mid, _ := ParseMessageID(msg.ReferID)
		next.Content.Refer = &MessageRefer{
			SessionID: sid,
			MessageID: mid,
		}
	}
	session.Messages = append(session.Messages, next)
	return MessageID(msg.SessionID, next.ID), nil
}

func (s *Server) GetOrCreatePrivateSessionID(u1, u2 string) (string, error) {
	user1 := s.Users[u1]
	user2 := s.Users[u2]
	if user1 == nil || user2 == nil {
		return "", errors.New("user not found")
	}
	for id, session := range s.Sessions {
		if !session.IsGroup() && slices.Contains(session.Object, u1) && slices.Contains(session.Object, u2) {
			return id, nil
		}
	}
	sid := fmt.Sprintf("30%d", len(s.Sessions))
	s.Sessions[sid] = &Session{
		Object:   []string{},
		Messages: make([]*Message, 0),
	}
	return sid, nil
}

func (s *Server) GetSessionIDByGroupID(id string) (string, error) {
	if s.Guilds[id] == nil {
		return "", errors.New("guild not found")
	}
	for sid, session := range s.Sessions {
		if session.IsGroup() && session.GroupID() == id {
			return sid, nil
		}
	}
	return "", errors.New("session not found")
}

func MessageID(sid, mid string) string {
	return fmt.Sprintf("%s@%s", sid, mid)
}

func ParseMessageID(mergedID string) (string, string, bool) {
	return strings.Cut(mergedID, "@")
}
