package network

import "net"

type ClientHandler struct {
	_dead int32
}

func (c *ClientHandler) start() {

}

func (c *ClientHandler) Close() {

}

func (c *ClientHandler) PostMessage(b []byte) error {
	return nil
}

func (c *ClientHandler) PostToMessage(b []byte, target net.Addr) error {
	return nil
}
