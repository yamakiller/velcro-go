package apps

import (
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/rpc/server"
)

func NewClientPools(s *server.RpcServer) server.RpcPool {
	return &ClientPools{pls: sync.Pool{
		New: func() interface{} {
			c := &TestClient{RpcClient: server.NewRpcClientConn(s)}
			c.Register(reflect.TypeOf(&protos.Auth{}), c.OnAuth)
			return c
		},
	}}
}

type ClientPools struct {
	pls sync.Pool
}

func (drp *ClientPools) Get() server.RpcClient {
	return drp.pls.Get().(server.RpcClient)
}

func (drp *ClientPools) Put(s server.RpcClient) {
	s.Destory()
	drp.pls.Put(s)
}
