package server

import (
	"sync"
)

type IRpcAllocator interface {
	Get() IRpcConnector
	Put(s IRpcConnector)
}

type defaultRpcAllocator struct {
	m sync.Pool
}

func (drp *defaultRpcAllocator) Get() IRpcConnector {
	return drp.m.Get().(IRpcConnector)
}

func (drp *defaultRpcAllocator) Put(c IRpcConnector) {
	c.Destory()
	drp.m.Put(c)
}
