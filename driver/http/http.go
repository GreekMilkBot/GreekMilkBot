package http

import "github.com/GreekMilkBot/GreekMilkBot/driver"

type HTTPDriver struct {
	driver.BaseDriver
}

func NewHttpDriver() *HTTPDriver {
	return &HTTPDriver{}
}
