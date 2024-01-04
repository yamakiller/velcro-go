package clientpool

import (

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
)

type IConnect interface {
	WithNode(intrusive.INode)
	WithAffiliation(perant *ConnectPool)
	WithTimeout(t int64)
	Timeout() int64
	Dial(string, time.Duration) error
	Redial() error
	RequsetMessage(message proto.Message, timeout int64) (interface{}, error)
	Ping() error
	Close() error
}


type BaseConnect struct {
	node intrusive.INode
	affiliation *ConnectPool
	timeout     int64
}

func (pc *BaseConnect) WithNode(node intrusive.INode) {
	pc.node = node
}

func (pc *BaseConnect) Node() intrusive.INode {
	return pc.node
}

func (pc *BaseConnect) WithAffiliation(aff *ConnectPool) {
	pc.affiliation = aff
}

func (pc *BaseConnect) Affiliation() *ConnectPool{
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
	value IConnect
}



