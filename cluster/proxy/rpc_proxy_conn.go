package proxy

import (
	"reflect"

	"github.com/yamakiller/velcro-go/rpc/client/asyn"
)

type RpcProxyConn struct {
	proxy *RpcProxy
	repe  *RpcProxyConnRepeat
	*asyn.Conn
}

func (rcc *RpcProxyConn) Close() {
	rcc.repe.stop()
	rcc.repe = nil
	rcc.Conn.Close()
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
	// 获取目标主机地址
	host := rpcx.ToAddress()
	// 取消平衡器中的目标主机
	rpcx.proxy.Lock()
	rpcx.proxy.alive[host] = false
	rpcx.proxy.Unlock()
	rpcx.proxy.balancer.Remove(host)
	// 如果重连器不为空,说明需要启动重连器.
	if rpcx.repe != nil {
		rpcx.repe.start()
	}

	rpcx.proxy.LogDebug("%s closed", rpcx.ToAddress())
}

func (rpcx *RpcProxyConn) Receive(msg interface{}) {
	if rpcx.proxy == nil || rpcx.proxy.recvice == nil {
		return
	}

	rpcx.proxy.recvice(msg)

	rpcx.proxy.LogDebug("receive %s", reflect.TypeOf(msg))
}
