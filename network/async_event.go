package network

import "net"

type AcceptMessage struct {
}

type PingMessage struct {
}

type RecviceMessage struct {
	Data []byte
	Addr net.Addr
}

type ClosedMessage struct {
}
