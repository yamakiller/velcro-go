package rpcserver

import (
	"sync"

	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
)

type RpcPool interface {
	Get() RpcClient
	Put(s RpcClient)
}

func NewDefaultRpcPool(s *RpcServer) RpcPool {
	return &defaultRpcPool{pls: sync.Pool{
		New: func() interface{} {
			return &RpcClientConn{
				recvice:         circbuf.New(32768, &syncx.NoMutex{}),
				methods:         make(map[interface{}]func(*RpcClientContext) interface{}),
				register:        s.Register,
				unregister:      s.UnRegister,
				unmarshal:       s.UnMarshal,
				marshalping:     s.MarshalPing,
				marshalresponse: s.MarshalResponse,
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
