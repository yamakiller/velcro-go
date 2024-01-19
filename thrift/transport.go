package thrift

import (
	"context"
	"errors"
	"io"
)

var errTransportInterrupted = errors.New("Transport Interrupted")

type Flusher interface {
	Flush() (err error)
}

type ContextFlusher interface {
	Flush(ctx context.Context) (err error)
}

type ReadSizeProvider interface {
	RemainingBytes() (num_bytes uint64)
}

// Encapsulates the I/O layer
type TTransport interface {
	io.ReadWriteCloser
	ContextFlusher
	ReadSizeProvider

	// Opens the transport for communication
	Open() error

	// Returns true if the transport is open
	IsOpen() bool
}

type stringWriter interface {
	WriteString(s string) (n int, err error)
}

// 这是具有额外功能的"增强型"传输.您需要使用其中之一来构建协议.
// 值得注意的是, TSocket 并没有实现这个接口, 直接在协议中使用
// TSocket 总是错误的.
type TRichTransport interface {
	io.ReadWriter
	io.ByteReader
	io.ByteWriter
	stringWriter
	ContextFlusher
	ReadSizeProvider
}
