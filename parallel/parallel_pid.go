package parallel

import (
	"sync/atomic"
	"unsafe"
)

func NewPID(address, Id string) *PID {
	return &PID{
		Address: address,
		Id:      Id,
	}
}

func (pid *PID) ref(parallelSystem *ParallelSystem) Handler {
	p := (*Handler)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&pid.h))))
	if p != nil {
		if l, ok := (*p).(*ParallelHandler); ok && atomic.LoadInt32(&l._death) == 1 {
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&pid.h)), nil)
		} else {
			return *p
		}
	}

	ref, exists := parallelSystem._handlers.Get(pid)
	if exists {
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&pid.h)), unsafe.Pointer(&ref))
	}

	return ref
}

func (pid *PID) postUsrMessage(parallelSystem *ParallelSystem, message interface{}) {
	pid.ref(parallelSystem).postUsrMessage(pid, message)
}

func (pid *PID) postSysMessage(parallelSystem *ParallelSystem, message interface{}) {
	pid.ref(parallelSystem).postSysMessage(pid, message)
}

// Equal 两个 PID 是否相等
func (pid *PID) Equal(other *PID) bool {
	if pid != nil && other == nil {
		return false
	}

	return pid.Id == other.Id && pid.Address == other.Address
}

func (pid *PID) ToString() string {
	return pid.Address + pid.Id
}
