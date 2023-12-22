package service

import (
	cmap "github.com/orcaman/concurrent-map"
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

	s := &Service{groups: newGroupMap()}
	s.RpcServer = rpcserver.New(rpcserver.WithPool(NewServiceClientPool(s)))

	return s
}

func newGroupMap() *groupMap {
	sm := &groupMap{}
	sm.tags = make([]tagMap, 32)

	for i := 0; i < len(sm.tags); i++ {
		sm.tags[i].cmap = cmap.New()
	}

	return sm
}

type Service struct {
	*rpcserver.RpcServer
	// ---- 成员变量----
	groups *groupMap // 服务分组
}

type tagMap struct {
	tag  string
	cmap cmap.ConcurrentMap
}

type groupMap struct {
	tags []tagMap
}

func (gm *groupMap) getBucket(tag string) cmap.ConcurrentMap {
	//hash := murmur32.Sum32([]byte(tag))
	//index := uint32(hash) & (uint32(len(gm.tags)) - 1)
	//return gm.tags[index]

}

func (s *Service) RegisetrGroup(vaddr, tag string, clientId *network.ClientID) {
	bucket := s.groups.getBucket(tag)

	bucket.SetIfAbsent(vaddr, clientId)
}
func (hr *HandleRegistryValue) Push(handler Handler, id string) (*ClientID, bool) {
	bucket := hr._localCIDs.getBucket(id)

	return &ClientID{
		Address: hr.Address,
		Id:      id,
	}, bucket.SetIfAbsent(id, handler)
}
