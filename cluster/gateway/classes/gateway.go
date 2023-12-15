package gateway

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/network"
)

// Gateway 网关
type Gateway struct {
	intelServerNetwork *network.NetworkSystem
	logger             logs.LogAgent
}

func (p *Gateway) Start(s service.Service) error {
	p.logger = produceLogger()
	p.intelServerNetwork = network.NewTCPServerNetworkSystem(
		network.WithMetricProviders(nil),
		network.WithNetworkTimeout(2000),
		network.WithProducer(newLinker),
		network.WithLoggerFactory(p.loggerFactory),
	)

	return nil
}

func (g *Gateway) loggerFactory(system *network.NetworkSystem) logs.LogAgent {
	return g.logger
}
