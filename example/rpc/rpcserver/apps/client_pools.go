package apps

import (
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

func NewClientPools(s *rpcserver.RpcServer) rpcserver.RpcPool {
	return &ClientPools{pls: sync.Pool{
		New: func() interface{} {
			//return &RpcClient{parent: s, recvice: circbuf.New(32768, &syncx.NoMutex{})}

			c := &TestClient{RpcClient: rpcserver.NewRpcClient(s)}
			c.Register(reflect.TypeOf(&protos.Auth{}), c.OnAuth)
			return c
		},
	}}
}

type ClientPools struct {
	pls sync.Pool
}

func (drp *ClientPools) Get() network.Client {
	return drp.pls.Get().(network.Client)
}

func (drp *ClientPools) Put(s *rpcserver.RpcClient) {
	// s.clientID = nil
	// s.recvice.Reset()
	drp.pls.Put(s)
}
