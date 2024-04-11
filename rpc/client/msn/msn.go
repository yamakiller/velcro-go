package msn

import (
	"sync"
	"sync/atomic"
)

var (
	instance *Msn
	instonce sync.Once
)

func Instance() *Msn {
	instonce.Do(func() {
		instance = &Msn{id: 1}
	})

	return instance
}

type Msn struct {
	id int32
}

func (msn *Msn) NextId() int32 {
	for {
		if id := atomic.AddInt32(&msn.id, 1); id > 0 {
			return id
		} else if atomic.CompareAndSwapInt32(&msn.id, id, 1) {
			return 1
		}
	}
}
