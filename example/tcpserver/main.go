package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/tcpserver/apps"
)

func main() {
	guard := application.Guardian{
		Name:        "TestTCPServer",
		Display:     "TTCPServer",
		Description: "网络系统TCP服务系统测试",
	}

	o, err := guard.Startup(&apps.Program{})
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}
}
