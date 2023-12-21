package proxy

import (
	"reflect"

	"github.com/yamakiller/velcro-go/rpc/rpcclient"
)

type RpcProxyConn struct {
	proxy *RpcProxy
	*rpcclient.Conn
}

func (rpcx *RpcProxyConn) Connected() {
	if rpcx.proxy == nil || rpcx.proxy.connected == nil {
		return
	}

	rpcx.proxy.connected(rpcx)
}

func (rpcx *RpcProxyConn) Closed() {
	if rpcx.proxy == nil {
		return
	}

	rpcx.proxy.Lock()
	rpcx.proxy.alive[rpcx.ToAddress()] = false
	rpcx.proxy.Unlock()
	rpcx.proxy.balancer.Remove(rpcx.ToAddress())
	rpcx.proxy.LogDebug("%s closed", rpcx.ToAddress())
}

func (rpcx *RpcProxyConn) Receive(msg interface{}) {
	if rpcx.proxy == nil || rpcx.proxy.recvice == nil {
		return
	}

	rpcx.proxy.recvice(msg)

	rpcx.proxy.LogDebug("receive %s", reflect.TypeOf(msg))
}
