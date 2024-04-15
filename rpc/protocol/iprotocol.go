package protocol

import (
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type IProtocol interface {
	thrift.TProtocol
	GetBytes()[]byte
	Write(b[]byte) (int, error)
	Reader() circbuf.Reader
	Release()
	Close()
}