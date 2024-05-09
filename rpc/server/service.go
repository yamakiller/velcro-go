package server

import (
	"errors"
	"fmt"

	"github.com/yamakiller/velcro-go/rpc/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc/pkg/serviceinfo"
)

type service struct {
	svcInfo *serviceinfo.ServiceInfo
	handler interface{}
}

func newService(svcInfo *serviceinfo.ServiceInfo, handler interface{}) *service {
	return &service{svcInfo: svcInfo, handler: handler}
}

type services struct {
	svcSearchMap                       map[string]*service // key: "svcName.methodName" or "methodName", value: svcInfo
	svcMap                             map[string]*service // key: serivce name, value svcInfo
	conflictingMethodHasFallbackSvcMap map[string]bool
	fallbackSvc                        *service
}

func newServices() *services {
	return &services{
		svcSearchMap:                       map[string]*service{},
		svcMap:                             map[string]*service{},
		conflictingMethodHasFallbackSvcMap: map[string]bool{},
	}
}

func (s *services) addService(svcInfo *serviceinfo.ServiceInfo, handler interface{}, registerOpts *RegisterOptions) error {
	svc := newService(svcInfo, handler)

	if err := s.checkCombineServiceWithOtherService(svcInfo); err != nil {
		return err
	}

	if err := s.checkMultipleFallbackService(registerOpts, svc); err != nil {
		return err
	}

	s.svcMap[svcInfo.ServiceName] = svc
	s.createSearchMap(svcInfo, svc, registerOpts)
	return nil
}

// 注册组合服务时,不允许注册其他服务.
func (s *services) checkCombineServiceWithOtherService(svcInfo *serviceinfo.ServiceInfo) error {
	if len(s.svcMap) > 0 {
		if _, ok := s.svcMap["CombineService"]; ok || svcInfo.ServiceName == "CombineService" {
			return errors.New("only one service can be registered when registering combine service")
		}
	}
	return nil
}

func (s *services) checkMultipleFallbackService(registerOpts *RegisterOptions, svc *service) error {
	if registerOpts.IsFallbackService {
		if s.fallbackSvc != nil {
			return fmt.Errorf("multiple fallback services cannot be registered. [%s] is already registered as a fallback service", s.fallbackSvc.svcInfo.ServiceName)
		}
		s.fallbackSvc = svc
	}
	return nil
}

func (s *services) createSearchMap(svcInfo *serviceinfo.ServiceInfo, svc *service, registerOpts *RegisterOptions) {
	for methodName := range svcInfo.Methods {
		s.svcSearchMap[remote.BuildMultiServiceKey(svcInfo.ServiceName, methodName)] = svc
		if svcFromMap, ok := s.svcSearchMap[methodName]; ok {
			s.handleConflictingMethod(svcFromMap, svc, methodName, registerOpts)
		} else {
			s.svcSearchMap[methodName] = svc
		}
	}
}

func (s *services) handleConflictingMethod(svcFromMap, svc *service, methodName string, registerOpts *RegisterOptions) {
	s.registerConflictingMethodHasFallbackSvcMap(svcFromMap, methodName)
	s.updateWithFallbackSvc(registerOpts, svc, methodName)
}

func (s *services) registerConflictingMethodHasFallbackSvcMap(svcFromMap *service, methodName string) {
	if _, ok := s.conflictingMethodHasFallbackSvcMap[methodName]; !ok {
		if s.fallbackSvc != nil && svcFromMap.svcInfo.ServiceName == s.fallbackSvc.svcInfo.ServiceName {
			// svc which is already registered is a fallback service
			s.conflictingMethodHasFallbackSvcMap[methodName] = true
		} else {
			s.conflictingMethodHasFallbackSvcMap[methodName] = false
		}
	}
}

func (s *services) updateWithFallbackSvc(registerOpts *RegisterOptions, svc *service, methodName string) {
	if registerOpts.IsFallbackService {
		s.svcSearchMap[methodName] = svc
		s.conflictingMethodHasFallbackSvcMap[methodName] = true
	}
}

func (s *services) getSvcInfoMap() map[string]*serviceinfo.ServiceInfo {
	svcInfoMap := map[string]*serviceinfo.ServiceInfo{}
	for name, svc := range s.svcMap {
		svcInfoMap[name] = svc.svcInfo
	}
	return svcInfoMap
}

func (s *services) getSvcInfoSearchMap() map[string]*serviceinfo.ServiceInfo {
	svcInfoSearchMap := map[string]*serviceinfo.ServiceInfo{}
	for name, svc := range s.svcSearchMap {
		svcInfoSearchMap[name] = svc.svcInfo
	}
	return svcInfoSearchMap
}
