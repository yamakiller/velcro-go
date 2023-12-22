package proxy

// RpcProxyStrategyOneToOne 单点请求
type RpcProxyStrategyOneToOne struct {
}

func (rpx *RpcProxyStrategyOneToOne) RequestMessage(conn *RpcProxyConn, message interface{}, timeout int64) (interface{}, error) {
	return conn.RequestMessage(message, uint64(timeout))
}
