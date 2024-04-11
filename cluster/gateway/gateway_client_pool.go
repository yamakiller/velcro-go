package gateway

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	// "github.com/yamakiller/velcro-go/utils/circbuf"
)

type GatewayClientPool interface {
	Get() network.Client
	Put(c Client)
}

func NewDefaultGatewayClientPool(g *Gateway) GatewayClientPool {
	return &defaultGatewayClientPool{pls: sync.Pool{
		New: func() interface{} {
			if g == nil {
				return nil
			}
			return &ClientConn{gateway: g,
				// requestTimeout: request_timeout,
				recvice: circbuf.NewLinkBuffer(4096),
				proto:   protocol.NewBinaryProtocol()}
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
