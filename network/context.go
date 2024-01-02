package network

import "net"

const (
	stateAccept int32 = iota
	stateAlive
	//stateClosing
	stateClosed
)

type Context interface {
	// Client
	Client() Client
	// Self clientId
	Self() *ClientID
	// NetworkSystem System object
	NetworkSystem() *NetworkSystem
	// Message 返回当前数据
	Message() []byte
	MessageFrom() net.Addr
	PostMessage(*ClientID, []byte) error
	PostToMessage(*ClientID, []byte, net.Addr) error
	// Close 关闭当前连接
	Close(*ClientID)
}
