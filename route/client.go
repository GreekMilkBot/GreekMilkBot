package route

import "errors"

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

}

func Receive(expr) {

}
