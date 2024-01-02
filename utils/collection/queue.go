package collection

import "sync"

func NewQueue(cap int, mutex sync.Locker) *Queue {
	return &Queue{
		_cap:               cap,
		_overloadThreshold: cap * 2,
		_buffer:            make([]interface{}, cap),
		_mu:                mutex,
	}
}

// Queue 连续可膨胀队列
type Queue struct {
	_cap  int
	_head int
	_tail int

	_overload          int
	_overloadThreshold int

	_buffer []interface{}
	_mu     sync.Locker
}

func (qe *Queue) Destory() {
	qe._mu.Lock()
	defer qe._mu.Unlock()

	qe._head = 0
	qe._tail = 0

	qe._buffer = make([]interface{}, 0)
}

// Push Insert an object
// @Param (interface{}) item
func (qe *Queue) Push(item interface{}) {
	qe._mu.Lock()
	defer qe._mu.Unlock()
	qe.unpush(item)
}

func (qe *Queue) Next() (interface{}, bool) {
	qe._mu.Lock()
	defer qe._mu.Unlock()
	return qe.unnext()
}

// Pop doc
// @Method Pop @Summary Take an object, If empty return nil
// @Return (interface{}) return object
// @Return (bool)
func (qe *Queue) Pop() (interface{}, bool) {
	qe._mu.Lock()
	defer qe._mu.Unlock()
	return qe.unpop()
}

// Overload Detecting queues exceeding the limit [mainly used for warning records]
// @Return (int)
func (qe *Queue) Overload() int {
	if qe._overload != 0 {
		overload := qe._overload
		qe._overload = 0
		return overload
	}
	return 0
}

// Length Length of the Queue
// @Return (int) length
func (qe *Queue) Length() int {
	var (
		head int
		tail int
		cap  int
	)
	qe._mu.Lock()
	head = qe._head
	tail = qe._tail
	cap = qe._cap
	qe._mu.Unlock()

	if head <= tail {
		return tail - head
	}
	return tail + cap - head
}

func (qe *Queue) unpush(item interface{}) {
	//utils.AssertEmpty(item, "error push is nil")
	qe._buffer[qe._tail] = item
	qe._tail++
	if qe._tail >= qe._cap {
		qe._tail = 0
	}

	if qe._head == qe._tail {
		qe.expand()
	}
}

func (qe *Queue) unpop() (interface{}, bool) {
	var resultSucces bool
	var result interface{}
	if qe._head != qe._tail {
		resultSucces = true
		result = qe._buffer[qe._head]
		qe._buffer[qe._head] = nil
		qe._head++
		if qe._head >= qe._cap {
			qe._head = 0
		}

		length := qe._tail - qe._tail
		if length < 0 {
			length += qe._cap
		}
		for length > qe._overloadThreshold {
			qe._overload = length
			qe._overloadThreshold *= 2
		}
	}

	return result, resultSucces
}

func (qe *Queue) unnext() (interface{}, bool) {
	var resultSucces bool
	var result interface{}
	if qe._head != qe._tail {
		result = qe._buffer[qe._head]
		resultSucces = true
	}

	return result, resultSucces
}

func (qe *Queue) expand() {
	newBuff := make([]interface{}, qe._cap*2)
	for i := 0; i < qe._cap; i++ {
		newBuff[i] = qe._buffer[(qe._head+i)%qe._cap]
	}

	qe._head = 0
	qe._tail = qe._cap
	qe._cap *= 2

	qe._buffer = newBuff
}
