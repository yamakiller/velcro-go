package server

import (
	"errors"
	"fmt"
	"sync"

	internal_server "github.com/yamakiller/velcro-go/rpc2/internal/server"
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote/remotesvr"
	"github.com/yamakiller/velcro-go/rpc2/pkg/serviceinfo"
	"github.com/yamakiller/velcro-go/utils/endpoint"
)

type Server interface {
	RegisterService(svcInfo *serviceinfo.ServiceInfo, handler interface{}, opts ...RegisterOption) error
	GetServiceInfos() map[string]*serviceinfo.ServiceInfo
	Run() error
	Stop() error
}

type server struct {
	opt           *internal_server.Options
	svcs          *services
	targetSvcInfo *serviceinfo.ServiceInfo

	// actual rpc service 实现 biz
	eps     endpoint.Endpoint
	mws     []endpoint.Middleware
	svr     remotesvr.Server
	stopped sync.Once
	isRun   bool

	sync.Mutex
}

func NewServer(opts ...Option) Server {
	s := &server{
		opt:  internal_server.NewOptions(opts),
		svcs: newServices(),
	}
	s.init()
	return s
}

func (s *server) init() {
}

// RegisterService should not be called by users directly.
func (s *server) RegisterService(svcInfo *serviceinfo.ServiceInfo, handler interface{}, opts ...RegisterOption) error {
	s.Lock()
	defer s.Unlock()

	return nil
}

func (s *server) GetServiceInfos() map[string]*serviceinfo.ServiceInfo {
	return s.svcs.getSvcInfoSearchMap()
}

// Run runs the server.
func (s *server) Run() (err error) {
	s.Lock()
	s.isRun = true
	s.Unlock()

	if err = s.check(); err != nil {
		return err
	}
	s.findAndSetDefaultService()
	//
	svrCfg := s.opt.Remote
	return nil
}

func (s *server) Stop() (err error) {
	s.stopped.Do(func() {
		s.Lock()
		defer s.Unlock()

		muShutdownHooks.Lock()
		for i := range onShutdown {
			onShutdown[i]()
		}
		muShutdownHooks.Unlock()

		if s.opt.RegistryInfo != nil {
			err = s.opt.Registry.Unregister(s.opt.RegistryInfo)
			s.opt.RegistryInfo = nil
		}
		if s.svr != nil {
			if e := s.svr.Stop(); e != nil {
				err = e
			}
			s.svr = nil
		}
	})

	return
}

func (s *server) check() error {
	if len(s.svcs.svcMap) == 0 {
		return errors.New("run: no service, Use RegisterService to set one ")
	}

	return checkFallbackServiceForConflictingMethods(s.svcs.conflictingMethodHasFallbackSvcMap, s.opt.RefuseTrafficWithoutServiceName)
}

func checkFallbackServiceForConflictingMethods(conflictingMethodHasFallbackSvcMap map[string]bool, refuseTrafficWithoutServiceName bool) error {
	if refuseTrafficWithoutServiceName {
		return nil
	}
	for name, hasFallbackSvc := range conflictingMethodHasFallbackSvcMap {
		if !hasFallbackSvc {
			return fmt.Errorf("method name [%s] is conflicted between services but no fallback service is specified", name)
		}
	}
	return nil
}

func (s *server) findAndSetDefaultService() {
	if len(s.svcs.svcMap) == 1 {
		s.targetSvcInfo = getDefaultSvcInfo(s.svcs)
	}
}

// getDefaultSvc is used to get one ServiceInfo from map
func getDefaultSvcInfo(svcs *services) *serviceinfo.ServiceInfo {
	if len(svcs.svcMap) > 1 && svcs.fallbackSvc != nil {
		return svcs.fallbackSvc.svcInfo
	}
	for _, svc := range svcs.svcMap {
		return svc.svcInfo
	}
	return nil
}
