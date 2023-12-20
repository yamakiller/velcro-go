package proxy

import "github.com/yamakiller/velcro-go/rpc/rpcclient"

type RpcProxyConn struct {
	proxy *RpcProxy
	*rpcclient.Conn
}

func (rpcx *RpcProxyConn) Closed() {
	if rpcx.proxy == nil {
		return
	}

	rpcx.proxy.Lock()
	rpcx.proxy.alive[rpcx.ToAddress()] = false
	rpcx.proxy.Unlock()
	rpcx.proxy.balancer.Remove(rpcx.ToAddress())
}

func (rpcx *RpcProxyConn) Receive(msg interface{}) {
	if rpcx.proxy == nil || rpcx.proxy.recvice == nil {
		return
	}

	rpcx.proxy.recvice(msg)
}
