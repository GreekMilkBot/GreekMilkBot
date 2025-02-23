package http

import "github.com/GreekMilkBot/GreekMilkBot/driver"

type WebSockDriver struct {
	driver.BaseDriver
}

func NewWebSockDriver() *WebSockDriver {
	return &WebSockDriver{}
}
