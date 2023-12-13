package uclient

import "github.com/yamakiller/velcro-go/network"

func NewTestClient(system *network.NetworkSystem) network.Client {
	return &TestClient{}
}

type TestClient struct {
}

func (tc *TestClient) Accept(ctx network.Context) {
	ctx.Info("一个连接接入")
}

func (tc *TestClient) Recvice(ctx network.Context) {
	ctx.Info("接收到数据:%d\n", len(ctx.Message()))
	ctx.PostMessage(ctx.Self(), []byte("1212323"))
}

func (tc *TestClient) Closed(ctx network.Context) {
	ctx.Info("连接被关闭")
}
