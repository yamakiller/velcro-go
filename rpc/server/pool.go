package server

import (
	"sync"

	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type RpcPool interface {
	Get() RpcClient
	Put(s RpcClient)
}

func NewDefaultRpcPool(s *RpcServer) RpcPool {
	return &defaultRpcPool{pls: sync.Pool{
		New: func() interface{} {
			return &RpcClientConn{
				recvice:    circbuf.NewLinkBuffer(32),
				register:   s.Register,
				unregister: s.UnRegister,
			}

		},
	}}
}

type defaultRpcPool struct {
	pls sync.Pool
}

func (drp *defaultRpcPool) Get() RpcClient {
	return drp.pls.Get().(RpcClient)
}

func (drp *defaultRpcPool) Put(s RpcClient) {
	s.Destory()
	drp.pls.Put(s)
}
