package event

import "sync"

// Queue is a ring to collect events.
type Queue interface {
	Push(e *Event)
	Dump() interface{}
}

// NewQueue 创建一个固定大小的队列
func NewQueue(cap int) Queue {
	q := &queue{
		array:    make([]*Event, cap),
		tailFlag: make(map[uint32]*uint32, cap),
	}
	for i := 0; i <= cap; i++ {
		t := uint32(0)
		q.tailFlag[uint32(i)] = &t
	}
	return q
}

// queue 一个实现固定大小的队列
type queue struct {
	array    []*Event
	tail     uint32
	tailFlag map[uint32]*uint32
	mutex    sync.RWMutex
}

// Push 插入一个事件.
func (q *queue) Push(e *Event) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.array[q.tail] = e

	newVersion := (*(q.tailFlag[q.tail])) + 1
	q.tailFlag[q.tail] = &newVersion

	q.tail = (q.tail + 1) % uint32(len(q.array))
}

// Dump 以相反的顺序转储之前推送的事件.
func (q *queue) Dump() interface{} {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	results := make([]*Event, 0, len(q.array))
	pos := int32(q.tail)
	for i := 0; i < len(q.array); i++ {
		pos--
		if pos < 0 {
			pos = int32(len(q.array) - 1)
		}

		e := q.array[pos]
		if e == nil {
			return results
		}

		results = append(results, e)
	}

	return results
}
