package server

import (
	"sync"

	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type RpcPool interface {
	Get() RpcClient
	Put(s RpcClient)
}

func NewDefaultRpcPool(s *RpcServer) RpcPool {
	return &defaultRpcPool{pls: sync.Pool{
		New: func() interface{} {
			return &RpcClientConn{
				recvice:    circbuf.New(32768, &syncx.NoMutex{}),
				methods:    make(map[interface{}]func(*RpcClientContext) (protoreflect.ProtoMessage, error)),
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
