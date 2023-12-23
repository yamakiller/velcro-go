package service

import (
	cmap "github.com/orcaman/concurrent-map"
	murmur32 "github.com/twmb/murmur3"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/server"
)

func New(options ...ServiceConfigOption) *Service {
	// config := Configure(options...)

	s := &Service{groups: newSliceMap()}
	s.RpcServer = server.New(server.WithPool(NewServiceClientPool(s, s.RegisetrGroup, s.UnregisterGroup)))

	return s
}

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

type Service struct {
	*server.RpcServer
	// ---- 成员变量----
	groups *sliceMap
}

func (s *Service) RegisetrGroup(key string, clientId *network.ClientID) {
	bucket := s.groups.getBucket(key)
	bucket.Set(key, clientId)
}

func (s *Service) UnregisterGroup(key string, clientId *network.ClientID) {
	bucket := s.groups.getBucket(key)
	bucket.Remove(key)
}
