package client

import (
	"github.com/yamakiller/velcro-go/cluster/service"
	"github.com/yamakiller/velcro-go/network"
)

func NewLoginClient(system *network.NetworkSystem) network.Client {
	return &LoginClient{
		ServiceClient: &service.ServiceClient{},
	}
}

type LoginClient struct {
	*service.ServiceClient
}

func (lc *LoginClient) Closed(ctx network.Context) {
	lc.ServiceClient.Closed(ctx)
}
