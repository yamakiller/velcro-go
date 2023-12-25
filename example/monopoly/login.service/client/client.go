package client

import (
	"github.com/yamakiller/velcro-go/cluster/service"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/network"
)

func NewLoginClient(system *network.NetworkSystem) network.Client {
	return &LoginClient{
		ServiceClient: &service.ServiceClient{},
	}
}

type LoginClient struct {
	
	*service.ServiceClient
	playerid uint32
}

func (lc *LoginClient) Closed(ctx network.Context){
	rds.RemUser(lc.playerid)
	lc.ServiceClient.Closed(ctx)
}