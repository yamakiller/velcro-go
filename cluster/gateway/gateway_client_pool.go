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

func NewDefaultGatewayClientPool(g *Gateway) GatewayClientPool {
	return &defaultGatewayClientPool{pls: sync.Pool{
		New: func() interface{} {
			if g == nil {
				return nil
			}
			conn := &ClientConn{gateway: g,
				recvice: circbuf.NewLinkBuffer(32),
			}
			
			if conn.gateway.encryption == nil {
				conn.message_agent = NewGatewayMessageAgent(conn)
			}else{
				conn.message_agent = NewPubkeyMessageAgent(conn)
			}
			return conn
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
