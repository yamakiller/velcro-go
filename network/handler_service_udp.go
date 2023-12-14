package network

import (
	"net"
	sync "sync"

	"github.com/yamakiller/velcro-go/containers"
)

type udpMsg struct {
	data []byte
	addr *net.UDPAddr
}

type udpClientHandler struct {
	_c         *net.UDPConn
	_wmail     *containers.Queue
	_wmailcond *sync.Cond
	//_keepalive      uint32
	//_keepaliveError uint8
	_invoker       MessageInvoker
	_senderStopped chan struct{}
	_started       int32
	_closed        int32
	ClientHandler
}

func (c *udpClientHandler) start() {
	c._wmailcond.L.Lock()
	if c._closed != 0 {
		c._wmailcond.L.Unlock()
		return
	}

	c._started = 1
	c._wmailcond.L.Unlock()

	c._wmailcond.Broadcast()
}

func (c *udpClientHandler) Close() {
	c._wmailcond.L.Lock()
	if c._closed != 0 {
		c._wmailcond.L.Unlock()
		return
	}

	c._wmail.Push(nil)
	c._wmailcond.L.Unlock()

	c._wmailcond.Broadcast()
}

func (c *udpClientHandler) PostMessage(b []byte) {
	panic("udp undefine post message")
}

func (c *udpClientHandler) PostToMessage(b []byte, target net.Addr) {

	udpTarget, ok := target.(*net.UDPAddr)
	if !ok {
		panic("udp post to message Please enter net.UDPAddr")
	}

	c._wmailcond.L.Lock()
	if c._closed != 0 && c._started == 1 {
		c._wmailcond.L.Unlock()
		return
	}
	c._wmail.Push(&udpMsg{data: b, addr: udpTarget})
	c._wmailcond.L.Unlock()
	c._wmailcond.Signal()
}
