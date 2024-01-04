package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/apps"
)

func main() {
	guard := application.Guardian{
		Name:        "TestRPCClient",
		Display:     "TRPCClient",
		Description: "网络系统RPC服务系统测试客户端",
	}

	p := &apps.Program{}
	o, err := guard.Startup(p)
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}
}
