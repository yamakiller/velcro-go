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
	PostMessage(*ClientID, []byte)
	PostToMessage(*ClientID, []byte, net.Addr)
	// Close 关闭当前连接
	Close(*ClientID)
	// 日志
	Info(sfmt string, args ...interface{})
	Debug(sfmt string, args ...interface{})
	Error(sfmt string, args ...interface{})
	Warning(sfmt string, args ...interface{})
	Fatal(sfmt string, args ...interface{})
	Panic(sfmt string, args ...interface{})
}
