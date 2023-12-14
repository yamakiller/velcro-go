package network

import (
	"net"
	sync "sync"

	"github.com/yamakiller/velcro-go/containers"
)

// ClientHandler TCP服务客户端处理程序
type tcpConnectorHandler struct {
	_c             net.Conn
	_wmail         *containers.Queue
	_wmailcond     *sync.Cond
	_invoker       MessageInvoker
	_senderStopped chan struct{}
	_started       int32
	_closed        int32
	ClientHandler
}

func (c *tcpConnectorHandler) start() {
	c._wmailcond.L.Lock()
	if c._closed != 0 {
		c._wmailcond.L.Unlock()
		return
	}

	c._started = 1
	c._wmailcond.L.Unlock()

	c._wmailcond.Broadcast()
}

func (c *tcpConnectorHandler) Close() {
	c._wmailcond.L.Lock()
	if c._closed != 0 {
		c._wmailcond.L.Unlock()
		return
	}

	c._wmail.Push(nil)
	c._wmailcond.L.Unlock()

	c._wmailcond.Broadcast()
}

func (c *tcpConnectorHandler) PostMessage(b []byte) {
	c._wmailcond.L.Lock()
	if c._closed != 0 && c._started == 1 {
		c._wmailcond.L.Unlock()
		return
	}
	c._wmail.Push(b)
	c._wmailcond.L.Unlock()
	c._wmailcond.Signal()
}

func (c *tcpConnectorHandler) PostToMessage(b []byte, target net.Addr) {
	panic("tcp connector undefine post to message")
}
