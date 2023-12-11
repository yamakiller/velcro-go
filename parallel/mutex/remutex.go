package mutex

import (
	"sync"

	"github.com/yamakiller/velcro-go/utils"
)

// ReMutex doc
// @Struct ReMutex @Summary Reentrant mutex
type ReMutex struct {
	_mutex *sync.Mutex
	_owner int
	_count int
}

// Width doc
// @Method Width @Summary Sync lock association reentrant lock
// @Param (*sync.Mutex) mutex object
func (slf *ReMutex) Width(m *sync.Mutex) {
	slf._mutex = m
}

// Lock doc
// @Method Lock @Summary locking
func (slf *ReMutex) Lock() {
	me := utils.GetCurrentGoroutineID()
	if slf._owner == me {
		slf._count++
		return
	}

	slf._mutex.Lock()
}

// Unlock doc
// @Method Unlock doc : unlocking
func (slf *ReMutex) Unlock() {
	utils.Assert(slf._owner == utils.GetCurrentGoroutineID(), "illegalMonitorStateError")
	if slf._count > 0 {
		slf._count--
	} else {
		slf._mutex.Unlock()
	}
}
