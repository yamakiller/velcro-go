package router

import "github.com/yamakiller/velcro-go/cluster/proxy"

// Router 路由器
type Router struct {
	VAddress string              // 路由虚拟地址
	Proxy    *proxy.RpcProxy     // 路由器网络通信器
	rules    int32               // 路由器授权角色表
	commands map[string]struct{} // 路由器命令表
}

func (r *Router) IsRulePass(rule int32) bool {
	return (r.rules & rule) != 0
}
