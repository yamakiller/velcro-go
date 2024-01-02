package network

import "net"

type MessageInvoker interface {
	invokerAccept()
	invokerRecvice([]byte, net.Addr)
	invokerPing()
	invokerClosed()
	invokerEscalateFailure(reason interface{}, message interface{})
}
