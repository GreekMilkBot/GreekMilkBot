package route

import (
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"sync"
	"time"
)

type PacketType int

type Packet struct {
	Source string `json:"src"` // 来源

	Target   string    `json:"dest"`   // 目标
	CreateAt time.Time `json:"create"` // 创建时间

	Stack []string `json:"stack"` // 包的路由信息,每次操作、响应数据包均需要附加之前的来源
	Meta  []string `json:"meta"`  // 包的元数据，如果当前数据包为另一个包的响应则需要带上之前包的元数据

	Content map[string]any `json:"data"` // 消息内容 (struct)
}

type Router struct {
	packets chan Packet
	lock    *sync.RWMutex
	done    chan struct{}
	clients map[string]*Client
	targets map[string][]string
}

func (r *Router) loop() {
	for {
		select {
		case <-r.done:
			return
		case p := <-r.packets:
			count := 0
			func() {
				r.lock.RLock()
				defer r.lock.RUnlock()
				for _, target := range r.targets[p.Target] {
					count += 1
					go func(p Packet, client *Client, target string) {
						defer func() {
							if err := recover(); err != nil {
								log.Error("target %s parse error %s", target, err)
							}
						}()
						p.Stack = append(p.Stack, target)
						client.packet(p.Target, p)
					}(p, r.clients[target], target)
				}
			}()
		}
	}
}

func NewRouter(cache int) *Router {
	r := &Router{
		packets: make(chan Packet, cache),
		clients: make(map[string]*Client),
		targets: make(map[string][]string),
		lock:    &sync.RWMutex{},
		done:    make(chan struct{}),
	}
	go r.loop()
	return r
}
