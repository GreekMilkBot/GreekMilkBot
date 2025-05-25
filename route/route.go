package route

import (
	"sync"
	"time"
)

type PacketType int

type Packet struct {
	Source string `json:"src"` // 来源

	Target   string    `json:"dest"`   // 目标
	CreateAt time.Time `json:"create"` // 创建时间

	Stack []string `json:"stack"` // 包的路由信息,每次操作此数据包均需要附加之前的来源

	Content string `json:"data"` // 消息内容
}

type Router struct {
	packets chan Packet
	lock    *sync.RWMutex
	done    chan struct{}

	filter[]
}

func (r *Router) loop() {
	for {
		select {
		case <-r.done:
			return
		case p := <-r.packets:

		}
	}
}

func NewRouter(cache int) *Router {
	r := &Router{
		packets: make(chan Packet, cache),
		lock:    &sync.RWMutex{},
		done:    make(chan struct{}),
	}
	go r.loop()
	return r
}
