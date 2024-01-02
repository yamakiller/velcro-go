package memory

import (
	"sync"
)

const maxSize = 46

var caches [maxSize]sync.Pool

func init() {
	for i := 0; i < maxSize; i++ {
		size := 1 << i
		caches[i].New = func() interface{} {
			buf := make([]byte, 0, size)
			return buf
		}
	}
}

// 计算索引
func calcIndex(size int) int {
	if size == 0 {
		return 0
	}
	if isPowerOfTwo(size) {
		return bsr(size)
	}
	return bsr(size) + 1
}

// 分配一块内存
func Malloc(size int, capacity ...int) []byte {
	if len(capacity) > 1 {
		panic("too many arguments to Malloc")
	}
	var c = size
	if len(capacity) > 0 && capacity[0] > size {
		c = capacity[0]
	}
	var ret = caches[calcIndex(c)].Get().([]byte)
	ret = ret[:size]
	return ret
}

// 释放一块内存
func Free(buf []byte) {
	size := cap(buf)
	if !isPowerOfTwo(size) {
		return
	}
	buf = buf[:0]
	caches[bsr(size)].Put(buf)
}
