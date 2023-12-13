package network

import (
	"fmt"
	"net"
	"sync/atomic"
	"unsafe"
)

func (cid *ClientID) ref(system *NetworkSystem) Handler {
	p := (*Handler)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cid.h))))
	fmt.Printf("%+v\n", p)
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

func (cid *ClientID) PostMessage(system *NetworkSystem, message []byte) {
	cid.ref(system).PostMessage(message)
}

func (cid *ClientID) PostToMessage(system *NetworkSystem, message []byte, target net.Addr) {
	cid.ref(system).PostToMessage(message, target)
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
