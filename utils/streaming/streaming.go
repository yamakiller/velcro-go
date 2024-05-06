package streaming

import (
	"context"
	"io"
)

type Stream interface {
	// Context the stream context.Context
	Context() context.Context
	// RecvMsg 接收来自对等方的消息将阻塞，直到出现错误或收到非并发安全消息.
	RecvMsg(interface{}) error
	// SendMsg 向对等方发送消息将阻塞,直到出现错误或有足够的缓冲区来发送非并发安全性.
	SendMsg(interface{}) error
	// SendMsg 不具有并发安全性
	io.Closer
}

// WithDoFinish 应在何时实现:
// (1) 您想将流包装在客户端中间件中,并且
// (2) 你想手动调用streaming.FinishStream(stream, error)来记录流的结束
// 注意：DoFinish应该是可重入的, 最好使用sync.Once.
type WithDoFinish interface {
	DoFinish(error)
}

// Args endpoint request
type Args struct {
	Stream Stream
}

// Result endpoint response
type Result struct {
	Stream Stream
}
