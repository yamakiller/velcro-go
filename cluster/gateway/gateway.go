package gateway

import (
	"sync"

	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"
)

func New(options ...GatewayConfigOption) *Gateway {

	config := Configure(options...)

	g := &Gateway{clients: make(map[network.CIDKEY]Client)}
	g.system = config.NewNetworkSystem(
		network.WithMetricProviders(config.MetricsProvider),
		network.WithLoggerFactory(config.LoggerFactory),
		network.WithNetworkTimeout(config.NetowkTimeout),
		network.WithProducer(g.newClient),
	)

	if config.ClientPool == nil {
		config.ClientPool = NewDefaultGatewayClientPool(g)
	}

	g.clientPool = config.ClientPool
	g.routeURI = config.RouterURI
	g.routeProxyFrequency = config.RouteProxyFrequency
	g.routeProxyDialTimeout = config.RouteProxyDialTimeout
	g.routeProxyKleepalive = config.RouteProxyKleepalive
	g.routeProxyAlgorithm = config.RouteProxyAlgorithm
	g.vaddr = config.VAddr
	g.laddr = config.LAddr
	return g
}

// Gateway 网关
type Gateway struct {
	vaddr  string // 网关局域网虚拟地址
	laddr  string // 网关监听地址
	system *network.NetworkSystem
	// 连接客户端
	clients    map[network.CIDKEY]Client
	mu         sync.Mutex
	clientPool GatewayClientPool
	// 编码
	encryption *Encryption
	// 路由
	routeURI              string
	routeProxyFrequency   int32
	routeProxyDialTimeout int32
	routeProxyKleepalive  int32
	routeProxyAlgorithm   string
	routeGroup            *router.RouterGroup
}

func (g *Gateway) Start() error {
	r, err := router.Loader(g.routeURI,
		router.WithFrequency(g.routeProxyFrequency),
		router.WithAlgorithm(g.routeProxyAlgorithm),
		router.WithDialTimeout(g.routeProxyDialTimeout),
		router.WithKleepalive(g.routeProxyKleepalive),
		router.WithReceiveCallback(g.onProxyRecvice),
	)

	if err != nil {
		return err
	}

	// 尝试连接服务
	g.routeGroup = r
	g.routeGroup.Open()
	// 打开监听
	if err := g.system.Open(g.laddr); err != nil {
		return err
	}

	return nil
}

func (g *Gateway) Stop() error {

	if g.system != nil {
		g.system.Info("Gateway Shutdown Closing client")

		g.mu.Lock()
		for _, c := range g.clients {
			c.ClientID().UserClose()
		}
		g.mu.Unlock()
		g.system.Info("Gateway Shutdown Closed client")
		g.system.Shutdown()
		g.system.Info("Gateway Shutdown")
		g.system = nil
	}

	if g.routeGroup != nil {
		g.routeGroup.Shutdown()
	}
	return nil
}

func (g *Gateway) Register(cid *network.ClientID, l Client) {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, ok := g.clients[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	l.referenceIncrement()
	g.clients[network.Key(cid)] = l
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

func (g *Gateway) onProxyRecvice(message interface{}) {
	// TODO: 代理推送过来的回调函数
}

func (g *Gateway) newClient(system *network.NetworkSystem) network.Client {
	return g.clientPool.Get()
}

func (g *Gateway) free(l Client) {
	l.Destory()
	g.clientPool.Put(l)
}
