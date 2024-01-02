package serve

/*import (
	cmap "github.com/orcaman/concurrent-map"
	murmur32 "github.com/twmb/murmur3"
)

func newSliceMap() *sliceMap {
	sm := &sliceMap{}
	sm.bucker = make([]cmap.ConcurrentMap, 32)

	for i := 0; i < len(sm.bucker); i++ {
		sm.bucker[i] = cmap.New()
	}

	return sm
}

type sliceMap struct {
	bucker []cmap.ConcurrentMap
}

func (s *sliceMap) getBucket(key string) cmap.ConcurrentMap {
	hash := murmur32.Sum32([]byte(key))
	index := uint32(hash) & (uint32(len(s.bucker)) - 1)
	return s.bucker[index]
}
*/
