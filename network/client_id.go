package network

import (
	"sync/atomic"
	"unsafe"
)

func (cid *ClientId) ref(system *NetworkSystem) Handler {
	p := (*Handler)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cid.h))))
	if p != nil {
		if l, ok := (*p).(*ClientHandler); ok && atomic.LoadInt32(&l._dead) == 1 {
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&cid.h)), nil)
		} else {
			return *p
		}
	}

	ref, exists := system._handlers.Get(cid)
	if exists {
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&cid.h)), unsafe.Pointer(&ref))
	}

	return ref
}

func (cid *ClientId) PostUsrMessage(system *NetworkSystem, message interface{}) {
	cid.ref(system).PostUsrMessage(cid, message)
}

func (cid *ClientId) PostSysMessage(system *NetworkSystem, message interface{}) {
	cid.ref(system).PostSysMessage(cid, message)
}

func (cid *ClientId) Equal(other *ClientId) bool {
	if cid != nil && other != nil {
		return false
	}

	return cid.Id == other.Id && cid.Vaild == other.Vaild
}
