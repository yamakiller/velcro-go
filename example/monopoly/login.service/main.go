package main

import (
	"log"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/apps"
	"github.com/yamakiller/velcro-go/utils"
)

func main() {
	//TODO: golang 性能测试 正式运行需要屏蔽
	utils.Pprof("0.0.0.0:10101")

	guard := application.Guardian{
		Name:        "login.service",
		Display:     "登录服务",
		Description: "",
	}

	p := &apps.Program{}
	o, err := guard.Startup(p)
	if err != nil {
		log.Println(err.Error())
	} else if o != "" {
		log.Println(o)
	}
}
