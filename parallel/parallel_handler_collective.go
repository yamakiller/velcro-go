package parallel

import (
	"sync/atomic"

	cmap "github.com/orcaman/concurrent-map"
	murmur32 "github.com/twmb/murmur3"
)

func newHandlerCollective(parallelSystem *ParallelSystem) *HandlerCollective {
	return &HandlerCollective{
		_parallelSystem: parallelSystem,
		_address:        localAddress,
		_localPIDs:      newSliceMap(),
	}
}

func newSliceMap() *sliceMap {
	sm := &sliceMap{}
	sm._localPIDs = make([]cmap.ConcurrentMap, 1024)

	for i := 0; i < len(sm._localPIDs); i++ {
		sm._localPIDs[i] = cmap.New()
	}

	return sm
}

// AddressResolver 用于解析远程 Parallel
type AddressResolver func(*PID) (Handler, bool)

type sliceMap struct {
	_localPIDs []cmap.ConcurrentMap
}

type HandlerCollective struct {
	_sequenceID     uint64
	_parallelSystem *ParallelSystem
	_address        string
	_localPIDs      *sliceMap
	_remoteHandlers []AddressResolver
}

func (s *sliceMap) getBucket(key string) cmap.ConcurrentMap {
	hash := murmur32.Sum32([]byte(key))
	index := int(hash) & (len(s._localPIDs) - 1)

	return s._localPIDs[index]
}

const (
	localAddress = "nonehost"
	digits       = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~+"
)

// uint64 转换到 字符串
func uint64ToId(u uint64) string {
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

func (hc *HandlerCollective) NextId() string {
	counter := atomic.AddUint64(&hc._sequenceID, 1)

	return uint64ToId(counter)
}

// Push 插入一个处理程序
func (hc *HandlerCollective) Push(handler Handler, id string) (*PID, bool) {
	bucket := hc._localPIDs.getBucket(id)

	return &PID{
		Address: hc._address,
		Id:      id,
	}, bucket.SetIfAbsent(id, handler)
}

// PushAddressResolver 插入一个远程解析程序
func (hc *HandlerCollective) PushAddressResolver(handler AddressResolver) {
	hc._remoteHandlers = append(hc._remoteHandlers, handler)
}

func (hc *HandlerCollective) Erase(pid *PID) {
	bucket := hc._localPIDs.getBucket(pid.Id)

	ref, _ := bucket.Pop(pid.Id)
	if l, ok := ref.(*ParallelHandler); ok {
		atomic.StoreInt32(&l._death, 1)
	}
}

func (hc *HandlerCollective) Get(pid *PID) (Handler, bool) {
	if pid == nil {
		return hc._parallelSystem._deadLetter, false
	}

	if pid.Address != localAddress && pid.Address != hc._address {
		for _, handler := range hc._remoteHandlers {
			ref, ok := handler(pid)
			if ok {
				return ref, true
			}
		}

		return hc._parallelSystem._deadLetter, false
	}

	bucket := hc._localPIDs.getBucket(pid.Id)
	ref, ok := bucket.Get(pid.Id)

	if !ok {
		return hc._parallelSystem._deadLetter, false
	}

	return ref.(Handler), true
}

func (hc *HandlerCollective) GetLocal(tag string) (Handler, bool) {
	bucket := hc._localPIDs.getBucket(tag)
	ref, ok := bucket.Get(tag)

	if !ok {
		return hc._parallelSystem._deadLetter, false
	}

	return ref.(Handler), true
}
