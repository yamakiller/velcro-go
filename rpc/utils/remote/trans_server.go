package remote

import (
	"net"

	"github.com/yamakiller/velcro-go/utils/syncx"
)

// TransServerFactory is used to create TransServer instances.
type TransServerFactory interface {
	NewTransServer(opt *ServerOption, transHdlr ServerTransHandler) TransServer
}

// TransServer is the abstraction for remote server.
type TransServer interface {
	CreateListener(net.Addr) (net.Listener, error)
	BootstrapServer(net.Listener) (err error)
	Shutdown() error
	ConnCount() syncx.AtomicInt
}
