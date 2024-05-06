package detection

import (
	"context"
	"net"

	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/utils/endpoint"
	"github.com/yamakiller/velcro-go/vlog"
)

var noopHandler = noopSvrTransHandler{}

type noopSvrTransHandler struct{}

func (noopSvrTransHandler) Write(ctx context.Context, conn net.Conn, send remote.Message) (nctx context.Context, err error) {
	return ctx, nil
}

func (noopSvrTransHandler) Read(ctx context.Context, conn net.Conn, msg remote.Message) (nctx context.Context, err error) {
	return ctx, nil
}

func (noopSvrTransHandler) OnRead(ctx context.Context, conn net.Conn) error {
	return nil
}
func (noopSvrTransHandler) OnInactive(ctx context.Context, conn net.Conn) {}
func (noopSvrTransHandler) OnError(ctx context.Context, err error, conn net.Conn) {
	if conn != nil {
		vlog.ContextErrorf(ctx, "VELCRO: processing error, remoteAddr=%v, error=%s", conn.RemoteAddr(), err.Error())
	} else {
		vlog.ContextErrorf(ctx, "VELCRO: processing error, error=%s", err.Error())
	}
}

func (noopSvrTransHandler) OnMessage(ctx context.Context, args, result remote.Message) (context.Context, error) {
	return ctx, nil
}
func (noopSvrTransHandler) SetPipeline(pipeline *remote.TransPipeline)       {}
func (noopSvrTransHandler) SetInvokeHandleFunc(inkHdlFunc endpoint.Endpoint) {}
func (noopSvrTransHandler) OnActive(ctx context.Context, conn net.Conn) (context.Context, error) {
	return ctx, nil
}
