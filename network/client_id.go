package network

import (
	"net"
	"sync/atomic"
	"unsafe"
)

func (cid *ClientID) ref(system *NetworkSystem) Handler {
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

func (cid *ClientID) postMessage(system *NetworkSystem, message []byte) {
	cid.ref(system).postMessage(message)
}

func (cid *ClientID) postToMessage(system *NetworkSystem, message []byte, target net.Addr) {
	cid.ref(system).postToMessage(message, target)
}

func (cid *ClientID) Close(system *NetworkSystem) {
	cid.ref(system).Close()
}

func (cid *ClientID) ToString() string {
	return cid.Address + cid.Id
}

func (cid *ClientID) Equal(other *ClientID) bool {
	if cid != nil && other != nil {
		return false
	}

	return cid.Id == other.Id && cid.Address == other.Address
}
