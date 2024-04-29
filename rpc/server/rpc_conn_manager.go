package server

import "github.com/yamakiller/velcro-go/network"

type IRpcManagerConnector interface {
	Register(*network.ClientID, IRpcConnector)
	Unregister(*network.ClientID)
}
