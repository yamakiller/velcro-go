package network

import (
	"errors"
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

	ref, exists := system.handlers.Get(cid)
	if exists {
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&cid.h)), unsafe.Pointer(&ref))
	}

	return ref
}

func (cid *ClientID) PostUserMessage(message []byte) error {
	if !cid.IsVaild() || cid.h == nil {
		return errors.New("ClientID: invalid")
	}

	(*cid.h).PostMessage(message)

	return nil
}

func (cid *ClientID) PostUserToMessage(message []byte, target net.Addr) error {
	if !cid.IsVaild() || cid.h == nil {
		return errors.New("ClientID: invalid")
	}

	(*cid.h).PostToMessage(message, target)

	return nil
}

func (cid *ClientID) UserClose() error {
	if !cid.IsVaild() || cid.h == nil {
		return errors.New("ClientID: invalid")
	}
	(*cid.h).Close()
	return nil
}

func (cid *ClientID) postMessage(system *NetworkSystem, message []byte) error {
	return cid.ref(system).PostMessage(message)
}

func (cid *ClientID) postToMessage(system *NetworkSystem, message []byte, target net.Addr) error {
	return cid.ref(system).PostToMessage(message, target)
}

func (cid *ClientID) close(system *NetworkSystem) {
	cid.ref(system).Close()
}

func (cid *ClientID) IsVaild() bool {
	return atomic.LoadInt32(&cid.vaild) == 0
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

func NewClientID(address, id string) *ClientID {
	return &ClientID{Address: address, Id: id}
}
