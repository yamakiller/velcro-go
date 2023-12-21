package gateway

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
)

type GatewayClientPool interface {
	Get() network.Client
	Put(c Client)
}

func NewDefaultGatewayClientPool(g *Gateway) GatewayClientPool {
	return &defaultGatewayClientPool{pls: sync.Pool{
		New: func() interface{} {
			return &ClientConn{gateway: g,
				recvice: circbuf.New(32768, &syncx.NoMutex{})}
		},
	}}
}

type defaultGatewayClientPool struct {
	pls sync.Pool
}

func (drp *defaultGatewayClientPool) Get() network.Client {
	return drp.pls.Get().(network.Client)
}

func (drp *defaultGatewayClientPool) Put(c Client) {
	c.Destory()
	drp.pls.Put(c)
}
