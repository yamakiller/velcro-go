package apps

import (
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

func NewClientPools(s *rpcserver.RpcServer) rpcserver.RpcPool {
	return &ClientPools{pls: sync.Pool{
		New: func() interface{} {
			c := &TestClient{RpcClient: rpcserver.NewRpcClient(s)}
			c.Register(reflect.TypeOf(&protos.Auth{}), c.OnAuth)
			return c
		},
	}}
}

type ClientPools struct {
	pls sync.Pool
}

func (drp *ClientPools) Get() rpcserver.RpcClient {
	return drp.pls.Get().(rpcserver.RpcClient)
}

func (drp *ClientPools) Put(s rpcserver.RpcClient) {
	s.Destory()
	drp.pls.Put(s)
}
