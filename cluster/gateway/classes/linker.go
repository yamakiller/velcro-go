package gateway

import "github.com/yamakiller/velcro-go/network"

const (
	NONE_AUTH_LEVEL = -1
)

func newLinker(system *network.NetworkSystem) network.Client {
	return &linker{}
}

type linker struct {
	Client *network.ClientID
}

func (l *linker) Accept(ctx network.Context) {
	ctx.Info("一个连接接入")
}

func (l *linker) Ping(ctx network.Context) {
	ctx.Info("请求心跳")
}

func (l *linker) Recvice(ctx network.Context) {
	ctx.Info("接收到数据:%d", len(ctx.Message()))
	ctx.PostMessage(ctx.Self(), []byte("1212323"))
}

func (l *linker) Closed(ctx network.Context) {
	ctx.Info("连接被关闭")
}
