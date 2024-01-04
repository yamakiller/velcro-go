package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/apps"
)

func main() {

	guard := application.Guardian{
		Name:        "battle.service",
		Display:     "对战服务",
		Description: "为战用户提供对战NAT服务",
	}

	p := &apps.Program{}
	o, err := guard.Startup(p)
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}
}
