package rpcserver

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
)

type ClientGroup struct {
	clients map[network.CIDKEY]*RpcClient
	mu      sync.Mutex
}
