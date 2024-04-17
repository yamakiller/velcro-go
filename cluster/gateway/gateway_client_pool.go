package gateway

import (
	"sync"

	"github.com/yamakiller/velcro-go/cluster/proxy/messageproxy"
	"github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils/circbuf"
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
			conn := &ClientConn{gateway: g,
				// requestTimeout: request_timeout,
				recvice: circbuf.NewLinkBuffer(4096),
			}
			repeat := messageproxy.NewRepeatMessageProxy()
			repeat.Register(protocol.MessageName(&pubs.RequestMessage{}),NewRequestMessageProxy(conn))
			repeat.Register(protocol.MessageName(&pubs.PingMsg{}),NewPingMessageProxy(conn))

			if conn.gateway.encryption == nil {
				conn.message_proxy = messageproxy.NewMessageProxy(repeat)
			}else{
				pubkey :=NewPubkeyMessageProxy(conn)
				conn.message_proxy = messageproxy.NewMessageProxy(pubkey,repeat)
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
