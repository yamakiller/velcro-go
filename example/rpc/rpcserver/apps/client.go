package apps

import (
	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

type TestClient struct {
	rpcserver.RpcClient
}

func (tc *TestClient) OnAuth(ctx *rpcserver.RpcClientContext) interface{} {
	requst := ctx.Message().(*protos.Auth)
	return requst
}
