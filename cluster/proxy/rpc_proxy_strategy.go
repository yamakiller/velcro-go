package proxy

type RpcProxyStrategy interface {
	RequestMessage(*RpcProxyConn, interface{}, int64) (interface{}, error)
}
