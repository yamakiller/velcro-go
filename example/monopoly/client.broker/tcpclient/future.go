package tcpclient

import (
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

type Future struct {
	sequenceID int32
	cond       *sync.Cond
	done       bool
	request    interface{}
	result     proto.Message
	err        error
	t          *time.Timer
}

// Error 错误信息
func (ref *Future) Error() error {
	return ref.err
}

// Result 结果
func (ref *Future) Result() proto.Message {
	return ref.result
}

func (ref *Future) Wait() {
	ref.cond.L.Lock()
	for !ref.done {
		ref.cond.Wait()
	}
	ref.cond.L.Unlock()
}