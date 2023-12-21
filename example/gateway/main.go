package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/gateway/apps"
)

func main() {
	guard := application.Guardian{
		Name:        "TestGetwayServer",
		Display:     "TGetwayServer",
		Description: "网络系统Getway服务系统测试",
	}

	p := &apps.Program{}
	o, err := guard.Startup(p)
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}

}
