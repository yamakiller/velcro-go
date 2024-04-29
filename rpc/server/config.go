package server

import (
	"github.com/yamakiller/velcro-go/network"
)

type ConnConfig struct {
	network.Config

	Kleepalive int32
	Allocator  IRpcAllocator
}

func defaultConfig() *ConnConfig {
	return &ConnConfig{
		Kleepalive: 10 * 1000,
		Allocator:  nil,
	}
}
