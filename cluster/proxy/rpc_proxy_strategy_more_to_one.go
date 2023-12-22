package proxy

import (
	"errors"

	"github.com/yamakiller/velcro-go/rpc/rpcclient"
)

type RpcProxyStrategyMoreToOne struct {
}

func (rpx *RpcProxyStrategyMoreToOne) RequestMessage(host string, conn *RpcProxyConn, message interface{}, timeout int64) (interface{}, error) {
	if !conn.IsConnected() {
		return nil, errors.New("The target service is not alive")
	}

	newConn := &RpcProxyConn{}
	newConn.Conn = rpcclient.NewConn(
		rpcclient.WithKleepalive(conn.Config.Kleepalive),
		rpcclient.WithConnected(nil),
		rpcclient.WithClosed(newConn.Closed),
		rpcclient.WithReceive(newConn.Receive))

	if err := newConn.Dial(host, conn.proxy.dialTimeouot); err != nil {
		return nil, err
	}

	defer newConn.Close()

	return newConn.RequestMessage(message, uint64(timeout))
}
