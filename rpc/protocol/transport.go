package protocol

import "github.com/apache/thrift/lib/go/thrift"

type Transport struct {
	thrift.TMemoryBuffer
}

func (p *Transport) IsOpen() bool {
	return true
}

func (p *Transport) Open() error {
	return nil
}