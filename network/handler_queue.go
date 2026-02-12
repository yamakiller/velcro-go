package network

import (
	"sync"

	"github.com/yamakiller/velcro-go/utils"
)

func spawnQueue(cap int) *queue {
	return &queue{
		_cap:               cap,
		_overloadThreshold: cap * 2,
		_buffer:            make([]interface{}, cap),
		_cond:              sync.NewCond(&sync.Mutex{}),
	}
}

type queue struct {
	_cap  int
	_head int
	_tail int

	_overload          int
	_overloadThreshold int

	_buffer []interface{}
	// _sync   sync.Mutex
	_cond *sync.Cond
}

// Push Insert an object
// @Param (interface{}) item
func (slf *queue) Push(item interface{}) {
	// slf._sync.Lock()
	// defer slf._sync.Unlock()
	slf._cond.L.Lock()
	defer slf._cond.L.Unlock()
	slf.unpush(item)
	slf._cond.Signal()
}

// Pop doc
// @Method Pop @Summary Take an object, If empty return nil
// @Return (interface{}) return object
// @Return (bool)
func (slf *queue) Pop() (interface{}, bool) {
	// slf._sync.Lock()
	// defer slf._sync.Unlock()
	slf._cond.L.Lock()
	defer slf._cond.L.Unlock()
	if slf.size() == 0 {
		slf._cond.Wait()
	}
	return slf.unpop()
}

// Overload Detecting queues exceeding the limit [mainly used for warning records]
// @Return (int)
func (slf *queue) Overload() int {
	if slf._overload != 0 {
		overload := slf._overload
		slf._overload = 0
		return overload
	}
	return 0
}

// Length Length of the queue
// @Return (int) length
func (slf *queue) Length() int {
	slf._cond.L.Lock()
	defer slf._cond.L.Unlock()
	return slf.size()
}
func (slf *queue) size() int {
	var (
		head int
		tail int
		cap  int
	)
	head = slf._head
	tail = slf._tail
	cap = slf._cap

	if head <= tail {
		return tail - head
	}
	return tail + cap - head
}
func (slf *queue) unpush(item interface{}) {
	utils.AssertEmpty(item, "error push is nil")
	slf._buffer[slf._tail] = item
	slf._tail++
	if slf._tail >= slf._cap {
		slf._tail = 0
	}

	if slf._head == slf._tail {
		slf.expand()
	}
}

func (slf *queue) unpop() (interface{}, bool) {
	var resultSucces bool
	var result interface{}
	if slf._head != slf._tail {
		resultSucces = true
		result = slf._buffer[slf._head]
		slf._buffer[slf._head] = nil
		slf._head++
		if slf._head >= slf._cap {
			slf._head = 0
		}

		length := slf._tail - slf._head
		if length < 0 {
			length += slf._cap
		}
		for length > slf._overloadThreshold {
			slf._overload = length
			slf._overloadThreshold *= 2
		}
	}

	return result, resultSucces
}

func (slf *queue) expand() {
	newBuff := make([]interface{}, slf._cap*2)
	for i := 0; i < slf._cap; i++ {
		newBuff[i] = slf._buffer[(slf._head+i)%slf._cap]
	}

	slf._head = 0
	slf._tail = slf._cap
	slf._cap *= 2

	slf._buffer = newBuff
}
