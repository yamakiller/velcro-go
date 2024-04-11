package network

import (
	cmap "github.com/orcaman/concurrent-map"
	murmur32 "github.com/twmb/murmur3"
	"github.com/yamakiller/velcro-go/utils/snowflakealien"
	"github.com/yamakiller/velcro-go/utils/syncx"
)

const (
	LocalAddress = "nonhost"
)

func NewHandlerRegistry(system *NetworkSystem, vaddr string) *HandleRegistryValue {

	hrv := &HandleRegistryValue{
		Address:    vaddr,
		_system:    system,
		_node:      snowflakealien.NewNode(&syncx.NoMutex{}),
		_localCIDs: newSliceMap(),
	}

	return hrv
}

func newSliceMap() *sliceMap {
	sm := &sliceMap{}
	sm.LocalCIDs = make([]cmap.ConcurrentMap, 1024)

	for i := 0; i < len(sm.LocalCIDs); i++ {
		sm.LocalCIDs[i] = cmap.New()
	}

	return sm
}

type sliceMap struct {
	LocalCIDs []cmap.ConcurrentMap
}

func (s *sliceMap) getBucket(key string) cmap.ConcurrentMap {
	hash := murmur32.Sum32([]byte(key))
	index := uint32(hash) & (uint32(len(s.LocalCIDs)) - 1)
	return s.LocalCIDs[index]
}

const (
	digits = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~+"
)

func int64ToId(u int64) string {
	var buf [13]byte
	i := 13
	// base is power of 2: use shifts and masks instead of / and %
	for u >= 64 {
		i--
		buf[i] = digits[uintptr(u)&0x3f]
		u >>= 6
	}
	// u < base
	i--
	buf[i] = digits[uintptr(u)]
	i--
	buf[i] = '$'

	return string(buf[i:])
}

type HandleRegistryValue struct {
	Address    string
	_system    *NetworkSystem
	_node      *snowflakealien.Node
	_localCIDs *sliceMap
}

func (hr *HandleRegistryValue) NextId() string {
	counter := int64(hr._node.Generate())

	return hr._system.ID + "/" + int64ToId(counter)
}

func (hr *HandleRegistryValue) Push(handler Handler, id string) (*ClientID, bool) {
	bucket := hr._localCIDs.getBucket(id)

	return &ClientID{
		Address: hr.Address,
		ID:      id,
	}, bucket.SetIfAbsent(id, handler)
}

func (hr *HandleRegistryValue) Remove(cid *ClientID) {
	bucket := hr._localCIDs.getBucket(cid.ID)

	_, _ = bucket.Pop(cid.ID)
}

func (hr *HandleRegistryValue) Get(cid *ClientID) (Handler, bool) {
	if cid == nil {
		return nil, false
	}

	bucket := hr._localCIDs.getBucket(cid.ID)
	ref, ok := bucket.Get(cid.ID)

	if !ok {
		return nil, false
	}

	return ref.(Handler), true
}

func (hr *HandleRegistryValue) GetLocal(id string) (Handler, bool) {
	bucket := hr._localCIDs.getBucket(id)
	ref, ok := bucket.Get(id)

	if !ok {
		return nil, false
	}

	return ref.(Handler), true
}
