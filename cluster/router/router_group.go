package router

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map"
	"google.golang.org/protobuf/proto"

	rpcmessage "github.com/yamakiller/velcro-go/rpc/messages"
)

type RouterGroup struct {
	routes []*Router
	cmaps  cmap.ConcurrentMap
}

// Open 打开路由组
func (rg *RouterGroup) Open() {
	for _, router := range rg.routes {
		router.Proxy.Open()
	}

}

// Shutdown 关闭路由组
func (rg *RouterGroup) Shutdown() {
	for _, router := range rg.routes {
		router.Proxy.Shutdown()
	}
	rg.routes = rg.routes[:0]
	rg.cmaps.Clear()
}

// Push 插入路由器
func (rg *RouterGroup) Push(router *Router) error {
	rg.routes = append(rg.routes, router)
	for scmd := range router.commands {
		if !rg.cmaps.SetIfAbsent(scmd, router) {
			return fmt.Errorf("router group push fail %s existed", scmd)
		}
	}

	return nil
}

// Broadcast 广播数据
func (rg *RouterGroup) Broadcast(message proto.Message, qos rpcmessage.RpcQos) {
	for _, router := range rg.routes {
		router.Proxy.PostMessage(message, qos)
	}
}

// Get 获取一个路由器
func (rg *RouterGroup) Get(command string) *Router {
	r, ok := rg.cmaps.Get(command)
	if !ok {
		return nil
	}

	return r.(*Router)
}
