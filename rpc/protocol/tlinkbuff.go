package protocol

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type TLinkBuffer struct {
	*circbuf.LinkBuffer
	size int
}

type TTLinkBufferTransportFactory struct {
	size int
}

func (p *TTLinkBufferTransportFactory) GetTransport(trans thrift.TTransport) (thrift.TTransport, error) {
	if trans != nil {
		t, ok := trans.(*TLinkBuffer)
		if ok && t.size > 0 {
			return NewTLinkBufferLen(t.size), nil
		}
	}
	return NewTLinkBufferLen(p.size), nil
}

func NewTLinkBufferTransportFactory(size int) *TTLinkBufferTransportFactory {
	return &TTLinkBufferTransportFactory{size: size}
}

func NewTLinkBuffer() *TLinkBuffer {
	return &TLinkBuffer{LinkBuffer: circbuf.NewLinkBuffer(32), size: 0}
}

func NewTLinkBufferLen(size int) *TLinkBuffer {
	return &TLinkBuffer{LinkBuffer: circbuf.NewLinkBuffer(size), size: size}
}

func (p *TLinkBuffer) IsOpen() bool {
	return true
}

func (p *TLinkBuffer) Open() error {
	return nil
}

func (p *TLinkBuffer) Close() error {
	return p.LinkBuffer.Close()
}

// Flushing a memory buffer is a no-op
func (p *TLinkBuffer) Flush(ctx context.Context) error {
	return p.LinkBuffer.Flush()
}

func (p *TLinkBuffer) RemainingBytes() (num_bytes uint64) {
	return uint64(p.MallocLen())
}

func (p *TLinkBuffer) Write(b []byte) (int,error){
	n  ,err := p.WriteBinary(b)
	if err !=nil{
		return n, err
	}
	return n,p.LinkBuffer.Flush()
}

func (p *TLinkBuffer) Read(b []byte) (n int, err error){
	return 0,nil
}