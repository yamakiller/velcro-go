package serve

import (
	"sync"

	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/files"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func New(options ...ServantConfigOption) *Servant {
	config := configure(options...)
	s := &Servant{
		Config:  config,
		clients: make(map[network.CIDKEY]*network.ClientID)}
	s.System = network.NewTCPSyncServerNetworkSystem(
		network.WithKleepalive(config.Kleepalive),
		network.WithProducer(s.spawConn),
		network.WithVAddr(config.Name+"@"+config.VAddr))
	return s
}

type Servant struct {
	System      *network.NetworkSystem
	Config      *ServantConfig
	clients     map[network.CIDKEY]*network.ClientID
	clientMutex sync.Mutex
	routeGroup  *router.RouterGroup
}

func (s *Servant) Start() error {
	if s.Config.Router.URI != "" {
		routeGroup, err := router.Loader(files.NewLocalPathFull(s.Config.Router.URI),
			s.Config.Router.Algorithm)
		if err != nil {
			return err
		}

		s.routeGroup = routeGroup
		s.routeGroup.Open()
	}

	if err := s.System.Open(s.Config.LAddr); err != nil {
		s.routeGroup.Shutdown()
		return err
	}

	return nil
}

func (s *Servant) Stop() error {
	if s.routeGroup != nil {
		s.routeGroup.Shutdown()
	}

	if s.System != nil {
		vlog.Infof("%s Shutdown Closing client", s.Config.Name)
		s.clientMutex.Lock()
		for _, clientId := range s.clients {
			clientId.UserClose()
		}
		s.clientMutex.Unlock()
		vlog.Infof("%s Shutdown Closed client", s.Config.Name)
		s.System.Shutdown()
		vlog.Infof("%s Shutdown", s.Config.Name)
		s.System = nil
	}

	s.routeGroup = nil

	return nil
}

// FindRouter 查询路由
func (s *Servant) FindRouter(message proto.Message) *router.Router {
	if s.routeGroup == nil {
		return nil
	}
	return s.routeGroup.Get(string(protoreflect.FullName(proto.MessageName(message))))
}

func (s *Servant) FindAddrRouter(addr string) *router.Router {
	if s.routeGroup == nil {
		return nil
	}

	return s.routeGroup.Find(addr)
}

func (s *Servant) spawConn(system *network.NetworkSystem) network.Client {
	return &ServantClientConn{
		Servant: s,
		recvice: circbuf.NewLinkBuffer(4096),
		events: make(map[interface{}]interface{}),
	}
}
