package rpcclient

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// 请求器
type Future struct {
	sequenceID int32
	cond       *sync.Cond
	done       bool
	result     interface{}
	err        error
	t          *time.Timer
}

func (ref *Future) wait() {
	ref.cond.L.Lock()
	for !ref.done {
		ref.cond.Wait()
	}
	ref.cond.L.Unlock()
}

func (ref *Future) Stop(slf *Conn) {
	ref.cond.L.Lock()
	if ref.done {
		ref.cond.L.Unlock()

		return
	}

	ref.done = true
	tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&ref.t))))

	if tp != nil {
		tp.Stop()
	}

	//移除
	slf.removeFuture(ref.sequenceID)

	ref.cond.L.Unlock()
	ref.cond.Signal()
}
