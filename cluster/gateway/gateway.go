package gateway

import (
	"errors"
	"sync"

	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
)

func New(options ...GatewayConfigOption) *Gateway {

	config := Configure(options...)

	g := &Gateway{clients: make(map[network.CIDKEY]Client)}
	g.System = config.NewNetworkSystem(
		network.WithKleepalive(config.Kleepalive),
		network.WithProducer(g.newClient),
		network.WithVAddr("Gateway@"+config.VAddr),
	)

	if config.ClientPool == nil {
		// config.ClientPool = NewDefaultGatewayClientPool(g, config.RequestDefautTimeout)
		config.ClientPool = NewDefaultGatewayClientPool(g)
	}

	g.servant = serve.New(
		serve.WithProducerActor(g.newServantActor),
		serve.WithName("Gateway"),
		serve.WithLAddr(config.LAddrServant),
		serve.WithVAddr(config.VAddr),
		serve.WithKleepalive(config.KleepaliveServant),
		serve.WithRoute(config.Router),
	)

	g.Config = config

	return g
}

// Gateway 网关
type Gateway struct {
	System *network.NetworkSystem
	Config *GatewayConfig
	// 客户端--------------------
	clients    map[network.CIDKEY]Client
	mu         sync.Mutex
	encryption *Encryption
	//---------------------------
	// 内部服务来连接
	servant *serve.Servant
}

func (g *Gateway) Start() error {

	if err := g.servant.Start(); err != nil {
		return err
	}

	// 打开监听
	if err := g.System.Open(g.Config.LAddr); err != nil {
		g.servant.Stop()

		return err
	}

	return nil
}

func (g *Gateway) Stop() error {

	if g.System != nil {
		vlog.Info("Gateway Shutdown Closing client")

		g.mu.Lock()
		for _, c := range g.clients {
			c.ClientID().UserClose()
		}
		g.mu.Unlock()
		vlog.Info("Gateway Shutdown Closed client")
		g.System.Shutdown()
		vlog.Info("Gateway Shutdown")
		g.System = nil
	}

	if g.servant != nil {
		g.servant.Stop()
	}

	return nil
}

func (g *Gateway) Register(cid *network.ClientID, l Client) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.clients) >= g.Config.OnlineOfNumber {
		return errors.New("gateway: client fulled")
	}

	_, ok := g.clients[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	l.referenceIncrement()
	g.clients[network.Key(cid)] = l

	return nil
}

func (g *Gateway) UnRegister(cid *network.ClientID) {

	g.mu.Lock()
	l, ok := g.clients[network.Key(cid)]
	if !ok {
		g.mu.Unlock()
		return
	}

	ref := l.referenceDecrement()
	delete(g.clients, network.Key(cid))

	g.mu.Unlock()

	if ref == 0 {
		g.free(l)
	}
}

func (g *Gateway) GetClient(cid *network.ClientID) Client {
	g.mu.Lock()
	defer g.mu.Unlock()

	l, ok := g.clients[network.Key(cid)]
	if !ok {
		return nil
	}

	l.referenceIncrement()
	return l
}

func (g *Gateway) ReleaseClient(c Client) {
	g.mu.Lock()
	ref := c.referenceDecrement()
	g.mu.Unlock()

	if ref == 0 {
		g.free(c)
	}
}

// FindRouter 查询路由
func (g *Gateway) FindRouter(message proto.Message) *router.Router {
	return g.servant.FindRouter(message)
}

// NewClientID 创建客户端连接者ID
func (g *Gateway) NewClientID(id string) *network.ClientID {
	return g.System.NewClientID(id)
}

func (g *Gateway) newServantActor(conn *serve.ServantClientConn) serve.ServantClientActor {
	actor := &GatewayServantActor{gateway: g}

	//conn.Register(&prvs.RequestGatewayPush{}, actor.onRequestGatewayPush)
	//conn.Register(&prvs.RequestGatewayAlterRule{}, actor.onRequestGatewayAlterRule)
	//conn.Register(&prvs.RequestGatewayCloseClient{}, actor.onRequestGatewayCloseClient)

	return actor
}

func (g *Gateway) newClient(System *network.NetworkSystem) network.Client {
	return g.Config.ClientPool.Get()
}

func (g *Gateway) free(l Client) {
	l.Destory()
	g.Config.ClientPool.Put(l)
}
