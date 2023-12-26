package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/apps"
)

func main() {
	guard := application.Guardian{
		Name:        "login.service",
		Display:     "login.service",
		Description: "login.service服务系统测试",
	}

	p := &apps.Program{}
	o, err := guard.Startup(p)
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}
}