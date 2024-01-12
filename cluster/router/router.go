package router

import (
	"github.com/yamakiller/velcro-go/cluster/proxy"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Router 路由器
type Router struct {
	Proxy    *proxy.RpcProxy  // 路由器网络通信器
	rules    int32            // 路由器授权角色表
	commands map[string]int64 // 路由器命令表
}

func (r *Router) IsRulePass(rule int32) bool {
	return (r.rules & rule) != 0
}

func (r *Router) GetMessageTimeout(message proto.Message) int64 {
	if timeout, ok := r.commands[string(protoreflect.FullName(proto.MessageName(message)))]; ok {
		return timeout
	}
	return 0
}
