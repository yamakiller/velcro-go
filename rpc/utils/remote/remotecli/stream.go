package remotecli

import (
	"context"

	"github.com/yamakiller/velcro-go/rpc/utils/remote"
	"github.com/yamakiller/velcro-go/rpc/utils/rpcinfo"
	"github.com/yamakiller/velcro-go/rpc/utils/streaming"
)

// NewStream create a client side stream
func NewStream(ctx context.Context, ri rpcinfo.RPCInfo, handler remote.ClientTransHandler, opt *remote.ClientOption) (streaming.Stream, error) {
	cm := NewConnWrapper(opt.ConnPool)
	var err error
	for _, shdlr := range opt.StreamingMetaHandlers {
		ctx, err = shdlr.OnConnectStream(ctx)
		if err != nil {
			return nil, err
		}
	}
	rawConn, err := cm.GetConn(ctx, opt.Dialer, ri)
	if err != nil {
		return nil, err
	}
	return nphttp2.NewStream(ctx, opt.SvcInfo, rawConn, handler), nil
}
