package proxy

import "github.com/yamakiller/velcro-go/rpc/client"

type RpcProxyConn struct {
	proxy *RpcProxy
	client.LongConnPool
}
