package service

import "github.com/yamakiller/velcro-go/rpc/rpcserver"

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

type Service struct {
	*rpcserver.RpcServer
}
