package proxy

type RpcProxyStrategy interface {
	RequestMessage(string, *RpcProxyConn, interface{}, int64) (interface{}, error)
}
