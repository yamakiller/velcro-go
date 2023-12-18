package rpcserver

import "github.com/yamakiller/velcro-go/network"

func NewClient(system *network.NetworkSystem) network.Client {
	return &RpcClient{}
}

type RpcClient struct {
	ClientID *network.ClientID
}

func (tc *RpcClient) Accept(ctx network.Context) {
	ctx.Info("一个连接接入")
}

func (tc *RpcClient) Ping(ctx network.Context) {
	ctx.Info("请求心跳")
}

func (tc *RpcClient) Recvice(ctx network.Context) {
	ctx.Info("接收到数据:%d", len(ctx.Message()))
	ctx.PostMessage(ctx.Self(), []byte("1212323"))
}

func (tc *RpcClient) Closed(ctx network.Context) {
	ctx.Info("连接被关闭")
}
