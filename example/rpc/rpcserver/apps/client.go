package apps

import (
	"context"

	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

type TestClient struct {
	*rpcserver.RpcClient
}

func (tc *TestClient) OnAuth(timeout context.Context, ctx network.Context, message interface{}) interface{} {
	requst := message.(*protos.Auth)
	return requst
}
