package route

import (
	"errors"
	"strings"
	"time"
)

type Client struct {
	name   string
	router *Router
}

func (r *Router) NewClient(name string) (*Client, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return nil, errors.New("empty name")
	}
	if strings.Contains(name, "/") {
		return nil, errors.New("invalid name")
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.clients[name] != nil {
		return nil, errors.New("client exists")
	}
	return &Client{
		name:   name,
		router: r,
	}, nil
}

func (c *Client) Send(target string, msg any) error {
	c.router.lock.RLock()
	defer c.router.lock.RUnlock()
	c.router.packets <- Packet{
		Source:   c.name,
		Target:   target,
		CreateAt: time.Now(),
		Stack:    make([]string, 0),
		Meta:     make([]string, 0),
		Content:  nil,
	}
	panic("not reached")
}

func (c *Client) packet(target string, p Packet) {

}
