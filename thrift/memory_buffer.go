package thrift

import (
	"bytes"
	"context"
)

// Memory buffer-based implementation of the TTransport interface.
type TMemoryBuffer struct {
	*bytes.Buffer
	size int
}

type TMemoryBufferTransportFactory struct {
	size int
}

func (p *TMemoryBufferTransportFactory) GetTransport(trans TTransport) (TTransport, error) {
	if trans != nil {
		t, ok := trans.(*TMemoryBuffer)
		if ok && t.size > 0 {
			return NewTMemoryBufferLen(t.size), nil
		}
	}
	return NewTMemoryBufferLen(p.size), nil
}

func NewTMemoryBufferTransportFactory(size int) *TMemoryBufferTransportFactory {
	return &TMemoryBufferTransportFactory{size: size}
}

func NewTMemoryBuffer() *TMemoryBuffer {
	return &TMemoryBuffer{Buffer: &bytes.Buffer{}, size: 0}
}

func NewTMemoryBufferLen(size int) *TMemoryBuffer {
	buf := make([]byte, 0, size)
	return &TMemoryBuffer{Buffer: bytes.NewBuffer(buf), size: size}
}

func (p *TMemoryBuffer) IsOpen() bool {
	return true
}

func (p *TMemoryBuffer) Open() error {
	return nil
}

func (p *TMemoryBuffer) Close() error {
	p.Buffer.Reset()
	return nil
}

// Flushing a memory buffer is a no-op
func (p *TMemoryBuffer) Flush(ctx context.Context) error {
	return nil
}

func (p *TMemoryBuffer) RemainingBytes() (num_bytes uint64) {
	return uint64(p.Buffer.Len())
}
