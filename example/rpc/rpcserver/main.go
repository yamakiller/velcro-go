package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/rpc/rpcserver/apps"
)

func main() {
	guard := application.Guardian{
		Name:        "TestRPCServer",
		Display:     "TRPCServer",
		Description: "网络系统RPC服务系统测试",
	}

	p := &apps.Program{}
	o, err := guard.Startup(p)
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}
}