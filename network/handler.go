package network

import "net"

type Handler interface {
	start()
	PostMessage([]byte)
	PostToMessage([]byte, net.Addr)
	Close()
}
