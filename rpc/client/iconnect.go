package client

import (
	"time"

	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type IConnectPool interface { //连接池
	Remove(intrusive.INode)
}

type IConnect interface {
	WithNode(intrusive.INode)
	Node() intrusive.INode
	WithAffiliation(IConnectPool)
	WithTimeout(int64)
	Timeout() int64
	Dial(string, time.Duration) error
	Redial() error
	RequestMessage(protoreflect.ProtoMessage, int64) (*Future, error)
	Close()
}


type BaseConnect struct {
	node        intrusive.INode
	affiliation IConnectPool
	timeout     int64
}

func (pc *BaseConnect) WithNode(node intrusive.INode) {
	pc.node = node
}

func (pc *BaseConnect) Node() intrusive.INode {
	return pc.node
}

func (pc *BaseConnect) WithAffiliation(aff IConnectPool) {
	pc.affiliation = aff
}

func (pc *BaseConnect) Affiliation() IConnectPool {
	return pc.affiliation
}

func (pc *BaseConnect) WithTimeout(t int64) {
	pc.timeout = t
}

func (pc *BaseConnect) Timeout() int64 {
	return pc.timeout
}

type PoolLinkNode struct {
	intrusive.LinkedNode
	Conn IConnect
}
