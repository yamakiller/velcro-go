package mutex

import (
	"github.com/yamakiller/velcro-go/utils"
)

// ReSpinLock doc
// @Struct ReSpinLock @Summary Re-entrant spin lock
type ReSpinLock struct {
	_mutex *SpinLock
	_owner int
	_count int
}

// Width doc
// @Method Width doc : Spinlock association reentrant spin lock
// @Param (*SpinLock) width spinlock object
func (slf *ReSpinLock) Width(m *SpinLock) {
	slf._mutex = m
}

// Trylock doc
// @Method Trylock doc : Try to lock if you fail to get the lock return failure will not try again
// @Return (bool)
func (slf *ReSpinLock) Trylock() bool {
	me := utils.GetCurrentGoroutineID()
	if slf._owner == me {
		slf._count++
		return true
	}

	return slf._mutex.Trylock()
}

// Lock doc
// @Method Lock @Summary locking
func (slf *ReSpinLock) Lock() {
	me := utils.GetCurrentGoroutineID()
	if slf._owner == me {
		slf._count++
		return
	}

	slf._mutex.Lock()
}

// Unlock doc
// @Method Unlock @Summary unlocking
func (slf *ReSpinLock) Unlock() {
	utils.Assert(slf._owner == utils.GetCurrentGoroutineID(), "illegalMonitorStateError")
	if slf._count > 0 {
		slf._count--
	} else {
		slf._mutex.Unlock()
	}
}
