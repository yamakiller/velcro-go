package network

import "net"

type Handler interface {
	start()
	PostMessage([]byte) error
	PostToMessage([]byte, net.Addr) error
	Close()
}
