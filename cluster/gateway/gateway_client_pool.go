package gateway

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type GatewayClientPool interface {
	Get() network.Client
	Put(c Client)
}

func NewDefaultGatewayClientPool(g *Gateway, request_timeout int64) GatewayClientPool {
	return &defaultGatewayClientPool{pls: sync.Pool{
		New: func() interface{} {
			return &ClientConn{gateway: g,
				requestTimeout: request_timeout,
				recvice:        circbuf.NewLinkBuffer(4096)}
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
