package network

import (
	"fmt"
	"net"
	sync "sync"

	"github.com/yamakiller/velcro-go/containers"
)

var _ Handler = &tcpClientHandler{}

type ClientHandler struct {
	_dead int32
}

func (c *ClientHandler) start() {

}

func (c *ClientHandler) Close() {

}

func (c *ClientHandler) postMessage(b []byte) {

}

func (c *ClientHandler) postToMessage(b []byte, target net.Addr) {

}

// ClientHandler 客户端处理程序
type tcpClientHandler struct {
	_c             net.Conn
	_wmail         *containers.Queue
	_wmailcond     sync.Cond
	_keepalive     uint32
	_invoker       MessageInvoker
	_senderStopped chan struct{}
	_started       int32
	_closed        int32
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

	fmt.Printf("send close\n")
	c._wmail.Push(nil)
	c._wmailcond.L.Unlock()

	c._wmailcond.Broadcast()
}

func (c *tcpClientHandler) postMessage(b []byte) {
	c._wmailcond.L.Lock()
	if c._closed != 0 && c._started == 1 {
		c._wmailcond.L.Unlock()
		return
	}
	c._wmail.Push(b)
	c._wmailcond.L.Unlock()
	c._wmailcond.Signal()
}

func (c *tcpClientHandler) postToMessage(b []byte, target net.Addr) {
	panic("tcp undefine post to message")
}