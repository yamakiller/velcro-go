package rpcserver

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
	lsync "github.com/yamakiller/velcro-go/sync"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type RpcPool interface {
	Get() network.Client
	Put(s *RpcClient)
}

func NewDefaultRpcPool(s *RpcServer) RpcPool {
	return &defaultRpcPool{pls: sync.Pool{
		New: func() interface{} {
			return &RpcClient{parent: s, recvice: circbuf.New(32768, &lsync.NoMutex{})}
		},
	}}
}

type defaultRpcPool struct {
	pls sync.Pool
}

func (drp *defaultRpcPool) Get() network.Client {
	return drp.pls.Get().(network.Client)
}

func (drp *defaultRpcPool) Put(s *RpcClient) {
	s.clientID = nil
	s.recvice.Reset()
	drp.pls.Put(s)
}
