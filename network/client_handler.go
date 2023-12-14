package network

import "net"

type ClientHandler struct {
	_dead int32
}

func (c *ClientHandler) start() {

}

func (c *ClientHandler) Close() {

}

func (c *ClientHandler) PostMessage(b []byte) {

}

func (c *ClientHandler) PostToMessage(b []byte, target net.Addr) {

}
