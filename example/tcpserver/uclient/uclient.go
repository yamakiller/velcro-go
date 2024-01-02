package uclient

import (
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/vlog"
)

func NewTestClient(system *network.NetworkSystem) network.Client {
	return &TestClient{}
}

type TestClient struct {
}

func (tc *TestClient) Accept(ctx network.Context) {
	vlog.Info("一个连接接入")
}

func (tc *TestClient) Ping(ctx network.Context) {
	vlog.Info("请求心跳")
}

func (tc *TestClient) Recvice(ctx network.Context) {
	vlog.Info("接收到数据:%d", len(ctx.Message()))
	ctx.PostMessage(ctx.Self(), []byte("1212323"))
}

func (tc *TestClient) Closed(ctx network.Context) {
	vlog.Info("连接被关闭")
}
