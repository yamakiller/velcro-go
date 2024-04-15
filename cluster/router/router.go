package router

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/cluster/proxy"
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

func (r *Router) GetMessageTimeout(name string) int64 {
	if timeout, ok := r.commands[name]; ok {
		return timeout
	}
	return 0
}

func (r *Router) Call(ctx context.Context, method string, args, result thrift.TStruct) (thrift.ResponseMeta, error){
	if _, err := r.Proxy.RequestMessage(args, method, r.GetMessageTimeout(method)); err != nil {
		// vlog.Errorf("request client %s closed to %s", args.Req.ClientID.ToString(), "")
		return thrift.ResponseMeta{},err
	}
	return thrift.ResponseMeta{},nil
}