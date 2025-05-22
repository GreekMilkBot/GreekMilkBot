package bot

import (
	"container/list"
	"time"
)

type Packet struct {
	From string `json:"from"` // 来源 (必须是一个真实的注册位置)
	To   string `json:"to"`   // 目标 (任意内容，消息可被丢弃)

	Created time.Time `json:"created"` // 创建时间

	Content string `json:"content"` // 消息内容
}

type Router struct {
	list *list.List
}

func NewRouter() *Router {
	return &Router{
		list: list.New(),
	}
}
