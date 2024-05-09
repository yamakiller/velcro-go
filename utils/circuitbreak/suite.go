package circuitbreak

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/rpc/pkg/discovery"
	"github.com/yamakiller/velcro-go/rpc/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/utils/endpoint"
	"github.com/yamakiller/velcro-go/utils/event"
	"github.com/yamakiller/velcro-go/utils/verrors"
)

const (
	serviceCircuitBreakKey  = "service"
	instanceCircuitBreakKey = "instance"
	circuitBreakConfig      = "circuitbreak_config"
)

var defaultCBConfig = Config{Enable: true, ErrRate: 0.5, MinSample: 200}

// CircuitBreakConfig is policy config of CircuitBreaker.
// DON'T FORGET to update DeepCopy() and Equals() if you add new fields.
type Config struct {
	Enable    bool    `yaml:"enable"`
	ErrRate   float64 `yaml:"err_rate"`
	MinSample int64   `yaml:"min_sample"`
}

// DeepCopy returns a full copy of CircuitBreakConfig.
func (c *Config) DeepCopy() *Config {
	if c == nil {
		return nil
	}
	return &Config{
		Enable:    c.Enable,
		ErrRate:   c.ErrRate,
		MinSample: c.MinSample,
	}
}

func (c *Config) Equals(other *Config) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	return c.Enable == other.Enable && c.ErrRate == other.ErrRate && c.MinSample == other.MinSample
}

// GenServiceCircuitBreakKeyFunc 通过rpcinfo生成断路器密钥
// 您可以根据您的配置中心自定义配置键.
type GenServiceCircuitBreakKeyFunc func(ri rpcinfo.RPCInfo) string

type instanceCircuitBreakConfig struct {
	Config
	sync.RWMutex
}

// NewCircuitBreakSuite 创建一个新的Suite.
// 注意: 对于此版本中的每个客户端.
func NewSuite(genKey GenServiceCircuitBreakKeyFunc, options ...SuiteOption) *Suite {
	s := &Suite{
		genServiceCircuitBreakKey: genKey,
		instanceCircuitBreakConfig: instanceCircuitBreakConfig{
			Config: defaultCBConfig,
		},
		config: SuiteConfig{
			serviceGetErrorTypeFunc:  ErrorTypeOnServiceLevel,
			instanceGetErrorTypeFunc: ErrorTypeOnInstanceLevel,
		},
	}
	for _, option := range options {
		option(&s.config)
	}
	return s
}

type Suite struct {
	servicePanel    Panel
	serviceControl  *Control
	instancePanel   Panel
	instanceControl *Control

	genServiceCircuitBreakKey GenServiceCircuitBreakKeyFunc
	serviceCircuitBreakConfig sync.Map // map[serviceCircuitBreakKey]CircuitBreakConfig

	instanceCircuitBreakConfig instanceCircuitBreakConfig

	events event.Queue

	config SuiteConfig
}

func (s *Suite) ServiceMW() endpoint.Middleware {
	if s == nil {
		return endpoint.DummyMiddleware
	}
	s.initInstanceCircuitBreak()
	return NewCircuitBreakerMW(*s.serviceControl, s.servicePanel)
}

// InstanceCBMW 回一个新的实例级别 CircuitBreakerMW.
func (s *Suite) InstanceMW() endpoint.Middleware {
	if s == nil {
		return endpoint.DummyMiddleware
	}
	s.initInstanceCircuitBreak()
	return NewCircuitBreakerMW(*s.instanceControl, s.instancePanel)
}

// ServicePanel 返回circuitbreak Panel of service
func (s *Suite) ServicePanel() Panel {
	if s.servicePanel == nil {
		s.initService()
	}
	return s.servicePanel
}

// return circuitbreak Control of service
func (s *Suite) ServiceControl() *Control {
	if s.serviceControl == nil {
		s.initService()
	}
	return s.serviceControl
}

//	UpdateServiceConfig 更新服务CircuitBreaker配置.
//
// 建议在远程配置模块中调用该函数.
func (s *Suite) UpdateServiceConfig(key string, cfg Config) {
	s.serviceCircuitBreakConfig.Store(key, cfg)
}

// UpdateInstanceConfig 更新CircuitBreaker全局配置参数
func (s *Suite) UpdateInstanceConfig(cfg Config) {
	s.instanceCircuitBreakConfig.Lock()
	s.instanceCircuitBreakConfig.Config = cfg
	s.instanceCircuitBreakConfig.Unlock()
}

// SetEventBusAndQueue CircuitBreaker与事件产生关联
func (s *Suite) SetEventBusAndQueue(bus event.Bus, events event.Queue) {
	s.events = events
	if bus != nil {
		bus.Watch(discovery.ChangeEventName, s.discoveryChangeHandler)
	}
}

// Dump 是转储 CircuitBreaker 信息以进行调试查询.
func (s *Suite) Dump() interface{} {
	return map[string]interface{}{
		serviceCircuitBreakKey:  debugInfo(s.servicePanel),
		instanceCircuitBreakKey: debugInfo(s.instancePanel),
		circuitBreakConfig:      s.configInfo(),
	}
}

// Close Panel 释放关联资源
func (s *Suite) Close() error {
	if s.servicePanel != nil {
		s.servicePanel.Close()
		s.servicePanel = nil
		s.serviceControl = nil
	}
	if s.instancePanel != nil {
		s.instancePanel.Close()
		s.instancePanel = nil
		s.instanceControl = nil
	}
	return nil
}

func (s *Suite) initService() {
	if s.servicePanel != nil && s.serviceControl != nil {
		return
	}
	if s.genServiceCircuitBreakKey == nil {
		s.genServiceCircuitBreakKey = RPCInfo2Key
	}
	opts := Options{
		ShouldTripWithKey: s.svcTripFunc,
	}
	s.servicePanel, _ = NewPanel(s.onServiceStateChange, opts)

	svcKey := func(ctx context.Context, request interface{}) (string, bool) {
		rio := rpcinfo.GetRPCInfo(ctx)
		serviceCircuitBreakKey := s.genServiceCircuitBreakKey(rio)
		cfg, _ := s.serviceCircuitBreakConfig.LoadOrStore(serviceCircuitBreakKey, defaultCBConfig)
		enabled := cfg.(Config).Enable

		return serviceCircuitBreakKey, enabled
	}
	s.serviceControl = &Control{
		GetKey:       svcKey,
		GetErrorType: s.config.serviceGetErrorTypeFunc,
		DecorateError: func(ctx context.Context, request interface{}, err error) error {
			return verrors.ErrServiceCircuitBreak
		},
	}
}

func (s *Suite) initInstanceCircuitBreak() {
	if s.instancePanel != nil && s.instanceControl != nil {
		return
	}
	opts := Options{
		ShouldTripWithKey: s.insTripFunc,
	}
	s.instancePanel, _ = NewPanel(s.onInstanceStateChange, opts)

	instanceKey := func(ctx context.Context, request interface{}) (string, bool) {
		rio := rpcinfo.GetRPCInfo(ctx)
		instanceCircuitBreakKey := rio.To().Address().String()
		s.instanceCircuitBreakConfig.RLock()
		enabled := s.instanceCircuitBreakConfig.Enable
		s.instanceCircuitBreakConfig.RUnlock()
		return instanceCircuitBreakKey, enabled
	}
	s.instanceControl = &Control{
		GetKey:       instanceKey,
		GetErrorType: s.config.instanceGetErrorTypeFunc,
		DecorateError: func(ctx context.Context, request interface{}, err error) error {
			return verrors.ErrInstanceCircuitBreak
		},
	}
}

func (s *Suite) onStateChange(level, key string, oldState, newState State, m Metricer) {
	if s.events == nil {
		return
	}

	successes, failures, timeouts := m.Counts()
	var errRate float64
	if sum := successes + failures + timeouts; sum > 0 {
		errRate = float64(failures+timeouts) / float64(sum)
	}

	s.events.Push(&event.Event{
		Name: level + "_cbr",
		Time: time.Now(),
		Detail: fmt.Sprintf("%s: %s -> %s, (succ: %d, err: %d, timeout: %d, rate: %f)",
			key, oldState, newState, successes, failures, timeouts, errRate),
	})
}

func (s *Suite) onServiceStateChange(key string, oldState, newState State, m Metricer) {
	s.onStateChange(serviceCircuitBreakKey, key, oldState, newState, m)
}

func (s *Suite) onInstanceStateChange(key string, oldState, newState State, m Metricer) {
	s.onStateChange(instanceCircuitBreakKey, key, oldState, newState, m)
}

func (s *Suite) discoveryChangeHandler(e *event.Event) {
	if s.instancePanel == nil {
		return
	}
	extra := e.Addon.(*discovery.Change)
	for i := range extra.Removed {
		instCBKey := extra.Removed[i].Address().String()
		s.instancePanel.RemoveBreaker(instCBKey)
	}
}

func (s *Suite) svcTripFunc(key string) TripFunc {
	pi, _ := s.serviceCircuitBreakConfig.LoadOrStore(key, defaultCBConfig)
	p := pi.(Config)
	return RateTripFunc(p.ErrRate, p.MinSample)
}

func (s *Suite) insTripFunc(key string) TripFunc {
	s.instanceCircuitBreakConfig.RLock()
	errRate := s.instanceCircuitBreakConfig.ErrRate
	minSample := s.instanceCircuitBreakConfig.MinSample
	s.instanceCircuitBreakConfig.RUnlock()
	return RateTripFunc(errRate, minSample)
}

func (s *Suite) configInfo() map[string]interface{} {
	svcCBMap := make(map[string]interface{})
	s.serviceCircuitBreakConfig.Range(func(key, value interface{}) bool {
		svcCBMap[key.(string)] = value
		return true
	})
	s.instanceCircuitBreakConfig.RLock()
	instCBConfig := s.instanceCircuitBreakConfig.Config
	s.instanceCircuitBreakConfig.RUnlock()

	cbMap := make(map[string]interface{}, 2)
	cbMap[serviceCircuitBreakKey] = svcCBMap
	cbMap[instanceCircuitBreakKey] = instCBConfig
	return cbMap
}

// RPCInfo2Key is to generate circuit breaker key through rpcinfo
func RPCInfo2Key(ri rpcinfo.RPCInfo) string {
	if ri == nil {
		return ""
	}
	fromService := ri.From().ServiceName()
	toService := ri.To().ServiceName()
	method := ri.To().Method()

	sum := len(fromService) + len(toService) + len(method) + 2
	var buf strings.Builder
	buf.Grow(sum)
	buf.WriteString(fromService)
	buf.WriteByte('/')
	buf.WriteString(toService)
	buf.WriteByte('/')
	buf.WriteString(method)
	return buf.String()
}

func debugInfo(panel Panel) map[string]interface{} {
	dumper, ok := panel.(interface {
		DumpBreakers() map[string]Breaker
	})
	if !ok {
		return nil
	}
	cbMap := make(map[string]interface{})
	for key, breaker := range dumper.DumpBreakers() {
		cbState := breaker.State()
		if cbState == Closed {
			continue
		}
		cbMap[key] = map[string]interface{}{
			"state":             cbState,
			"successes in 10s":  breaker.Metricer().Successes(),
			"failures in 10s":   breaker.Metricer().Failures(),
			"timeouts in 10s":   breaker.Metricer().Timeouts(),
			"error rate in 10s": breaker.Metricer().ErrorRate(),
		}
	}
	if len(cbMap) == 0 {
		cbMap["msg"] = "all circuit breakers are in closed state"
	}
	return cbMap
}
