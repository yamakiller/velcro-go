package proxy

import (
	"github.com/yamakiller/velcro-go/rpc/client/clientpool"
)

type RpcProxyConn struct {
	proxy *RpcProxy
	*clientpool.ConnectPool
}

/*type RpcProxyConn struct {
	proxy *RpcProxy
	repe  *RpcProxyConnRepeat
	*client.Conn
}

func (rcc *RpcProxyConn) Close() {
	rcc.repe.stop()
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
	if !rpcx.proxy.isStopped() {
		rpcx.repe.start()
	}

	vlog.Errorf("%s closed", rpcx.ToAddress())
}*/
