package proxy

import "google.golang.org/protobuf/proto"

// RpcProxyStrategyOneToOne 单点请求
type RpcProxyStrategyOneToOne struct {
	proxy *RpcProxy
}

func (rpx *RpcProxyStrategyOneToOne) RequestMessage(host string, message proto.Message, timeout int64) (proto.Message, error) {
	return rpx.proxy.hostMap[host].RequestMessage(message, uint64(timeout))
}
