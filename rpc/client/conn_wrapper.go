package client

import (
	"net"
	"sync"
)

var connWrapperPool sync.Pool

func init() {
	connWrapperPool.New = newConnWrapper
}

// ConnWrapper wraps a connection.
type ConnWrapper struct {
	connPool ConnPool
	conn     net.Conn
}

// NewConnWrapper returns a new ConnWrapper using the given connPool and logger.
func NewConnWrapper(connPool ConnPool) *ConnWrapper {
	cm := connWrapperPool.Get().(*ConnWrapper)
	cm.connPool = connPool
	return cm
}

func newConnWrapper() interface{} {
	return &ConnWrapper{}
}

func (cm *ConnWrapper) zero() {
	cm.connPool = nil
	cm.conn = nil
}
