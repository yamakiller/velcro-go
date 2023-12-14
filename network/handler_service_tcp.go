package network

import (
	"net"
	sync "sync"

	"github.com/yamakiller/velcro-go/containers"
)

var _ Handler = &tcpClientHandler{}

// ClientHandler TCP服务客户端处理程序
type tcpClientHandler struct {
	_c              net.Conn
	_wmail          *containers.Queue
	_wmailcond      *sync.Cond
	_keepalive      uint32
	_keepaliveError uint8
	_invoker        MessageInvoker
	_senderStopped  chan struct{}
	_started        int32
	_closed         int32
	ClientHandler
}

func (c *tcpClientHandler) start() {
	c._wmailcond.L.Lock()
	if c._closed != 0 {
		c._wmailcond.L.Unlock()
		return
	}

	c._started = 1
	c._wmailcond.L.Unlock()

	c._wmailcond.Broadcast()
}

func (c *tcpClientHandler) Close() {
	c._wmailcond.L.Lock()
	if c._closed != 0 {
		c._wmailcond.L.Unlock()
		return
	}

	c._wmail.Push(nil)
	c._wmailcond.L.Unlock()

	c._wmailcond.Broadcast()
}

func (c *tcpClientHandler) PostMessage(b []byte) {
	c._wmailcond.L.Lock()
	if c._closed != 0 && c._started == 1 {
		c._wmailcond.L.Unlock()
		return
	}
	c._wmail.Push(b)
	c._wmailcond.L.Unlock()
	c._wmailcond.Signal()
}

func (c *tcpClientHandler) PostToMessage(b []byte, target net.Addr) {
	panic("tcp undefine post to message")
}
