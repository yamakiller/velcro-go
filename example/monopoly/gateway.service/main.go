package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/monopoly/gateway.service/apps"
	"github.com/yamakiller/velcro-go/utils"
)

func main() {

	//TODO: golang 性能测试 正式运行需要屏蔽
	utils.Pprof("0.0.0.0:10002")

	guard := application.Guardian{
		Name:        "gateway.service",
		Display:     "gateway.service",
		Description: "网络系统gateway.service服务系统测试",
	}

	p := &apps.Program{}
	o, err := guard.Startup(p)
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}
}
