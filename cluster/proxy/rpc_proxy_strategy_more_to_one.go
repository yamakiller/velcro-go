package proxy

import (
	"github.com/yamakiller/velcro-go/rpc/client/sync"
	"google.golang.org/protobuf/proto"
)

// RpcProxyStrategyMoreToOne 多点请求
type RpcProxyStrategyMoreToOne struct {
	proxy *RpcProxy
}

func (rpx *RpcProxyStrategyMoreToOne) RequestMessage(host string, message proto.Message, timeout int64) (proto.Message, error) {
	rpcc := sync.NewConn()
	if err := rpcc.Dial(host, rpx.proxy.dialTimeouot); err != nil {
		return nil, err
	}
	defer rpcc.Close()

	return rpcc.RequestMessage(message, uint64(timeout))
}
