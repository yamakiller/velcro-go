package proxy

import (
	"errors"

	"github.com/yamakiller/velcro-go/rpc/rpcsynclient"
)

// RpcProxyStrategyMoreToOne 多点请求
type RpcProxyStrategyMoreToOne struct {
}

func (rpx *RpcProxyStrategyMoreToOne) RequestMessage(conn *RpcProxyConn, message interface{}, timeout int64) (interface{}, error) {
	if !conn.IsConnected() {
		return nil, errors.New("The target service is not alive")
	}

	rpcc := rpcsynclient.NewConn(rpcsynclient.WithMarshalRequest(conn.Config.MarshalRequest),
		rpcsynclient.WithUnMarshal(conn.Config.UnMarshal))
	if err := rpcc.Dial(conn.ToAddress(), conn.proxy.dialTimeouot); err != nil {
		return nil, err
	}
	defer rpcc.Close()

	return rpcc.RequestMessage(message, uint64(timeout))
}
