package bot

import (
	"errors"
	"sync"
	"time"
)

type PacketType int

type Packet struct {
	ID    uint64   `json:"id"`    // 包的 id, 保证在每次会话中唯一
	Stack []string `json:"stack"` // 包的路由信息,每次操作此数据包均需要附加之前的来源

	Source   string    `json:"src"`    // 来源
	Target   string    `json:"dest"`   // 目标
	CreateAt time.Time `json:"create"` // 创建时间

	Content string `json:"data"` // 消息内容
}

type Router struct {
	packets chan Packet
	lock    *sync.RWMutex
}

func NewRouter(cache int) *Router {
	return &Router{
		packets: make(chan Packet, cache),
		lock:    &sync.RWMutex{},
	}
}

type Client struct {
	name   string
	router *Router
}

func (r *Router) NewClient(name string) (*Client, error) {
	if name == "" {
		return nil, errors.New("empty name")
	}
	return &Client{
		name:   name,
		router: r,
	}, nil
}

func (c *Client) Send(pType, target, msg string) error {

	panic("implement me")
}

func (c *Client) RPC() <-chan Packet {

}
