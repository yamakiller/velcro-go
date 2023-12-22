package proxy

type RpcProxyStrategyOneToOne struct {
}

func (rpx *RpcProxyStrategyOneToOne) RequestMessage(host string, conn *RpcProxyConn, message interface{}, timeout int64) (interface{}, error) {
	return conn.RequestMessage(message, uint64(timeout))
}
