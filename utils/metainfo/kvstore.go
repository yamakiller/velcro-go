package metainfo

import "sync"

type kvstore map[string]string

var kvpool sync.Pool

func newKVStore(size ...int) kvstore {
	kvs := kvpool.Get()
	if kvs == nil {
		if len(size) > 0 {
			return make(kvstore, size[0])
		}
		return make(kvstore)
	}
	return kvs.(kvstore)
}

func (store kvstore) size() int {
	return len(store)
}

func (store kvstore) recycle() {
	/*
			for k := range m {
				delete(m, k)
			}
		  ==>
			LEAQ    type.map[string]int(SB), AX
			MOVQ    AX, (SP)
			MOVQ    "".m(SB), AX
			MOVQ    AX, 8(SP)
			PCDATA  $1, $0
			CALL    runtime.mapclear(SB)
	*/
	for key := range store {
		delete(store, key)
	}
	kvpool.Put(store)
}
