package service

import (
	cmap "github.com/orcaman/concurrent-map"
	murmur32 "github.com/twmb/murmur3"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

/*func New(options ...ServiceConfigOption) *Service {
	// config := Configure(options...)

	s := &Service{
		_rpcs: rpcserver.New(),
		_registrys:make(map[network.CIDKEY]ServiceConn),
	}
	s._rpcs.WithPool(NewServiceClientPools(s,s._rpcs))
	return s
}

type ServiceConn struct{
	Vaddr string
	Tag string
}


type Service struct {
	_rpcs *rpcserver.RpcServer
	//服务注册
	_registrys map[network.CIDKEY]ServiceConn
	_mu         sync.Mutex

	clientPool ServiceClientPools
}

func (s *Service) Start(addr string) error {
	if err := s._rpcs.Open(addr); err != nil {
		return err
	}
	return nil
}
func (s *Service) Stop() error {
	s._rpcs.Shutdown()
	return nil
}
func (s *Service) Register(cid *network.ClientID, vaddr string,tag string) {
	s._mu.Lock()
	defer s._mu.Unlock()

	_, ok := s._registrys[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	s._registrys[network.Key(cid)] = ServiceConn{
		Vaddr: vaddr,
		Tag: tag,
	}
}

func (s *Service) UnRegister(cid *network.ClientID) {
	s._mu.Lock()
	_, ok := s._registrys[network.Key(cid)]
	if !ok {
		s._mu.Unlock()
		return
	}
	delete(s._registrys, network.Key(cid))
	s._mu.Unlock()
}*/

func New(options ...ServiceConfigOption) *Service {
	// config := Configure(options...)

	s := &Service{groups: newSliceMap()}
	s.RpcServer = rpcserver.New(rpcserver.WithPool(NewServiceClientPool(s)))

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
	*rpcserver.RpcServer
	// ---- 成员变量----
	groups *sliceMap
}

type tagMap struct {
	tag  string
	cmap cmap.ConcurrentMap
}

func (s *Service) RegisetrGroup(key string, clientId *network.ClientID) {
	bucket := s.groups.getBucket(key)
	bucket.Set(key, clientId)
}
