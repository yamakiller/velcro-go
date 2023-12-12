package network

import "net"

type Handler interface {
	start()
	postMessage([]byte)
	postToMessage([]byte, net.Addr)
	Close()
}
