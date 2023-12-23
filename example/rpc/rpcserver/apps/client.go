package apps

import (
	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/rpc/server"
)

type TestClient struct {
	server.RpcClient
}

func (tc *TestClient) OnAuth(ctx *server.RpcClientContext) interface{} {
	requst := ctx.Message().(*protos.Auth)
	return requst
}
