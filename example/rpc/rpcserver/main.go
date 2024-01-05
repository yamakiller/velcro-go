package main

import (
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/yamakiller/velcro-go/application"
	"github.com/yamakiller/velcro-go/example/rpc/rpcserver/apps"
)

func main() {

	go func() {
        http.HandleFunc("/debug/pprof/block", pprof.Index)
        http.HandleFunc("/debug/pprof/goroutine", pprof.Index)
        http.HandleFunc("/debug/pprof/heap", pprof.Index)
        http.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
 
        http.ListenAndServe("0.0.0.0:8889", nil)
    }()

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