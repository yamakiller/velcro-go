package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/behavior/apps"
)

func main() {
	guard := application.Guardian{
		Name:        "BehaviorServer",
		Display:     "BehaviorServer",
		Description: "网络系统Behavior服务系统测试",
	}

	p := &apps.Program{}
	o, err := guard.Startup(p)
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}

}
