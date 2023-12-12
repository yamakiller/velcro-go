package network

import "net"

type MessageInvoker interface {
	invokerAccept()
	invokerRecvice([]byte, *net.Addr)
	invokerClosed()
}
