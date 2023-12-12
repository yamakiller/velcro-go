package network

import (
	"time"

	"github.com/bwmarrin/snowflake"
	cmap "github.com/orcaman/concurrent-map"
	murmur32 "github.com/twmb/murmur3"
)

const (
	localAddress = "nonhost"
)

func NewHandlerRegistry(system *NetworkSystem) *HandleRegistryValue {
	snowflake.Epoch = time.Now().UnixNano() / 1e6
	node, _ := snowflake.NewNode(25768)

	hrv := &HandleRegistryValue{
		Address:    localAddress,
		_node:      node,
		_system:    system,
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
	index := uint32(hash)&uint32(len(s.LocalCIDs)) - 1

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
	_node      *snowflake.Node
	_system    *NetworkSystem
	_localCIDs *sliceMap
}

func (hr *HandleRegistryValue) NextId() string {
	counter := hr._node.Generate().Int64()

	return int64ToId(counter)
}

func (hr *HandleRegistryValue) Push(handler Handler, id string) (*ClientID, bool) {
	bucket := hr._localCIDs.getBucket(id)

	return &ClientID{
		Address: hr.Address,
		Id:      id,
	}, bucket.SetIfAbsent(id, handler)
}

func (hr *HandleRegistryValue) Remove(cid *ClientID) {
	bucket := hr._localCIDs.getBucket(cid.Id)

	_, _ = bucket.Pop(cid.Id)
}

func (hr *HandleRegistryValue) Get(cid *ClientID) (Handler, bool) {
	if cid == nil {
		return nil, false
	}

	bucket := hr._localCIDs.getBucket(cid.Id)
	ref, ok := bucket.Get(cid.Id)

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
