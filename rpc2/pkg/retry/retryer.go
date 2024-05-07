// Package retry 实现 rpc retry
package retry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/utils/circuitbreak"
	"github.com/yamakiller/velcro-go/vlog"
)

// ContainerOption is used when initializing a Container
type ContainerOption func(rc *Container)

// WithContainerCircuitBreakSuite specifies the CBSuite used in the retry circuitbreak
// retryer will use its ServiceControl and ServicePanel
// Its priority is lower than WithContainerCBControl and WithContainerCBPanel
func WithContainerCircuitBreakSuite(cbs *circuitbreak.Suite) ContainerOption {
	return func(rc *Container) {
		rc.container.suite = cbs
	}
}

// WithContainerCircuitBreakSuiteOptions specifies the circuitbreak.CBSuiteOption for initializing circuitbreak.CBSuite
func WithContainerCircuitBreakSuiteOptions(opts ...circuitbreak.SuiteOption) ContainerOption {
	return func(rc *Container) {
		rc.container.suiteOptions = opts
	}
}

// WithContainerCircuitBreakControl specifies the circuitbreak.Control used in the retry circuitbreaker
// It's user's responsibility to make sure it's paired with panel
func WithContainerCircuitBreakControl(ctrl *circuitbreak.Control) ContainerOption {
	return func(rc *Container) {
		rc.container.ctrl = ctrl
	}
}

// WithContainerCBPanel specifies the circuitbreaker.Panel used in the retry circuitbreaker
// It's user's responsibility to make sure it's paired with control
func WithContainerCircuitBreakPanel(panel circuitbreak.Panel) ContainerOption {
	return func(rc *Container) {
		rc.container.panel = panel
	}
}

// WithContainerCircuitBreakStat instructs the circuitbreak.RecordStat is called within the retryer
func WithContainerCircuitBreakStat() ContainerOption {
	return func(rc *Container) {
		rc.container.stat = true
	}
}

// WithContainerEnablePercentageLimit should be called for limiting the percentage of retry requests
func WithContainerEnablePercentageLimit() ContainerOption {
	return func(rc *Container) {
		rc.container.enablePercentageLimit = true
	}
}

// NewRetryContainerWithCircuitBreak 构建不进行断路器统计但获取统计结果的容器.在启用断路器的情况下使用:
// 例如:
//
//		cbs := circuitbreak.NewSuite(circuitbreak.RPCInfo2key)
//		retryContainer := retry.NewRetryContainerWithCircuitBreak(cbs.ServiceControl(), cbs.ServicePanel())
//	 	var opts []client.Option
//		opts = append(opts, client.WithRetryContainer(retryContainer))
//	    // enable service circuit breaker
//		opts = append(opts, client.WithMiddleware(cbs.ServiceMW()))
func NewRetryContainerWithCircuitBreak(cc *circuitbreak.Control, cp circuitbreak.Panel) *Container {
	return NewRetryContainer(WithContainerCircuitBreakControl(cc), WithContainerCircuitBreakPanel(cp))
}

// NewRetryer build a retryer with policy
func NewRetryer(p Policy, r *ShouldResultRetry, container *cbrContainer) (retryer Retryer, err error) {
	// just one retry policy can be enabled at same time
	if p.Type == BackupType {
		retryer, err = newBackupRetryer(p, container)
	} else {
		retryer, err = newFailureRetryer(p, r, container)
	}
	return
}

func newCircuitBreakSuite(opts []circuitbreak.SuiteOption) *circuitbreak.Suite {
	return circuitbreak.NewSuite(circuitbreak.RPCInfo2Key, opts...)
}

// RPCCallFunc is the definition with wrap rpc call
type RPCCallFunc func(context.Context, Retryer) (rpcinfo rpcinfo.RPCInfo, resp interface{}, err error)

// Retryer is the interface for Retry 实现
type Retryer interface {
	// AllowRetry 检查当前请求是否满足重试条件(例如:  circuit, retry times == 0, chain stop, dll).
	// 如果不满足, 则不会执行 Retryer.Do 并返回原因消息 无论是否能够重试,第一次仍执行.
	AllowRetry(ctx context.Context) (msg string, ok bool)
	// ShouldRetry 检查是否可以调用重试请求, 在 retryer.Do 中检查.
	// 如果不满足则返回原因信息
	ShouldRetry(ctx context.Context, err error, callTimes int, req interface{}, cbKey string) (string, bool)
	UpdatePolicy(policy Policy) error

	// Retry 策略执行函数. 返回的bool决定rio是否可以回收.
	Do(ctx context.Context, rpcCall RPCCallFunc, firstRIO rpcinfo.RPCInfo, request interface{}) (rpcinfo.RPCInfo, bool, error)
	AppendErrMsgIfNeeded(err error, ri rpcinfo.RPCInfo, msg string)

	// Prepare 在重试呼叫之前做一些需要做的事情
	Prepare(ctx context.Context, prevRI, retryRI rpcinfo.RPCInfo)
	Dump() map[string]interface{}
	Type() Type
}

func NewRetryContainer(opts ...ContainerOption) *Container {
	rc := &Container{
		container: &cbrContainer{
			suite: nil,
		},
		retryerMap: sync.Map{},
	}
	for _, opt := range opts {
		opt(rc)
	}

	if rc.container.enablePercentageLimit {
		// ignore cbSuite/cbCtl/cbPanel options
		rc.container = &cbrContainer{
			enablePercentageLimit: true,
			suite:                 newCircuitBreakSuite(rc.container.suiteOptions),
			suiteOptions:          rc.container.suiteOptions,
		}
	}

	container := rc.container
	if container.ctrl == nil && container.panel == nil {
		if container.suite == nil {
			container.suite = newCircuitBreakSuite(rc.container.suiteOptions)
			container.stat = true
		}
		container.ctrl = container.suite.ServiceControl()
		container.panel = container.suite.ServicePanel()
	}
	if !container.IsValid() {
		panic("VELCRO: invalid container")
	}
	return rc
}

type Container struct {
	isCodeCfg  bool
	retryerMap sync.Map // <method: retryer>
	container  *cbrContainer
	msg        string
	sync.RWMutex

	// shouldResultRetry 仅与 FailureRetry 一起使用
	shouldResultRetry *ShouldResultRetry
}

// Init to build Retryer with code config.
func (rc *Container) Init(mp map[string]Policy, rr *ShouldResultRetry) (err error) {
	// NotifyPolicyChange func may execute before Init func.
	// Because retry Container is built before Client init, NotifyPolicyChange can be triggered first
	rc.updateRetryer(rr)
	if err = rc.InitWithPolicies(mp); err != nil {
		return fmt.Errorf("NewRetryer in Init failed, err=%w", err)
	}
	return nil
}

// WithRetryIfNeeded to check if there is a retryer can be used and if current call can retry.
// When the retry condition is satisfied, use retryer to call
func (rc *Container) WithRetryIfNeeded(ctx context.Context, callOptRetry *Policy, rpcCall RPCCallFunc, ri rpcinfo.RPCInfo, request interface{}) (lastRI rpcinfo.RPCInfo, recycleRI bool, err error) {
	var retryer Retryer
	if callOptRetry != nil && callOptRetry.Enable {
		// build retryer for call level if retry policy is set up with callopt
		if retryer, err = NewRetryer(*callOptRetry, nil, rc.container); err != nil {
			vlog.Warnf("KITEX: new callopt retryer[%s] failed, err=%w", callOptRetry.Type, err)
		}
	} else {
		retryer = rc.getRetryer(ri)
	}

	// case 1(default, fast path): no retry policy
	if retryer == nil {
		if _, _, err = rpcCall(ctx, nil); err == nil {
			return ri, true, nil
		}
		return ri, false, err
	}

	// case 2: setup retry policy, but not satisfy retry condition eg: circuit, retry times == 0, chain stop, ddl
	if msg, ok := retryer.AllowRetry(ctx); !ok {
		if _, _, err = rpcCall(ctx, retryer); err == nil {
			return ri, true, err
		}
		if msg != "" {
			retryer.AppendErrMsgIfNeeded(err, ri, msg)
		}
		return ri, false, err
	}

	// case 3: do rpc call with retry policy
	lastRI, recycleRI, err = retryer.Do(ctx, rpcCall, ri, request)
	return
}

// Dump is used to show current retry policy
func (rc *Container) Dump() interface{} {
	rc.RLock()
	dm := make(map[string]interface{})
	dm["has_code_cfg"] = rc.isCodeCfg
	rc.retryerMap.Range(func(key, value interface{}) bool {
		if r, ok := value.(Retryer); ok {
			dm[key.(string)] = r.Dump()
		}
		return true
	})
	if rc.msg != "" {
		dm["msg"] = rc.msg
	}
	rc.RUnlock()
	return dm
}

// Close releases all possible resources referenced.
func (rc *Container) Close() (err error) {
	if rc.container != nil && rc.container.suite != nil {
		err = rc.container.suite.Close()
	}
	return
}

func (rc *Container) initRetryer(method string, p Policy) error {
	retryer, err := NewRetryer(p, rc.shouldResultRetry, rc.container)
	if err != nil {
		errMsg := fmt.Sprintf("new retryer[%s-%s] failed, err=%s, at %s", method, p.Type, err.Error(), time.Now())
		rc.msg = errMsg
		vlog.Warnf(errMsg)
		return err
	}

	rc.retryerMap.Store(method, retryer)
	if p.Enable {
		rc.msg = fmt.Sprintf("new retryer[%s-%s] at %s", method, retryer.Type(), time.Now())
	} else {
		rc.msg = fmt.Sprintf("disable retryer[%s-%s](enable=%t) %s", method, p.Type, p.Enable, time.Now())
	}
	return nil
}

func (rc *Container) updateRetryer(rr *ShouldResultRetry) {
	rc.Lock()
	defer rc.Unlock()

	rc.shouldResultRetry = rr
	if rc.shouldResultRetry != nil {
		rc.retryerMap.Range(func(key, value interface{}) bool {
			if fr, ok := value.(*failureRetryer); ok {
				fr.setSpecifiedResultRetryIfNeeded(rc.shouldResultRetry)
			}
			return true
		})
	}
}

func (rc *Container) getRetryer(ri rpcinfo.RPCInfo) Retryer {
	// the priority of specific method is high
	r, ok := rc.retryerMap.Load(ri.To().Method())
	if ok {
		return r.(Retryer)
	}
	r, ok = rc.retryerMap.Load(Wildcard)
	if ok {
		return r.(Retryer)
	}
	return nil
}

type cbrContainer struct {
	// In NewRetryContainer, 如果未设置 ctrl 和 panel,
	// Velcro 将使用 suite.ServiceControl() 和 suite.ServicePanel();如果 suite 为null，Velcro 将创建一个.
	suite *circuitbreak.Suite

	// 最好依赖suite自动创建
	ctrl  *circuitbreak.Control
	panel circuitbreak.Panel

	// stat和!enablePercentageLimit成立,则重试器将在rpcCall之后调用
	// circuitbreak.RecordStat来记录rpc失败/超时,以便在错误率超出阀值时间少重试请求.
	stat bool

	// 如果启用, Velcro 总是会创建一个circutibreak.Suite并让ctrl和panel使用,
	// 重试器会在rpcCall之前调用recordRetryStat,以精确控制重试请求占所有请求的百分比.
	enablePercentageLimit bool

	// 用于在 NewRetryContainer 内创建 circutibreak.Suite
	suiteOptions []circuitbreak.SuiteOption
}

// IsValid 判断 cbrContainer 是否被初始化(有效?)
func (c *cbrContainer) IsValid() bool {
	return c.ctrl != nil && c.panel != nil
}

func (rc *Container) InitWithPolicies(methodPolicies map[string]Policy) error {
	if methodPolicies == nil {
		return nil
	}
	rc.Lock()
	defer rc.Unlock()
	var inited bool
	for m := range methodPolicies {
		if methodPolicies[m].Enable {
			inited = true
			if _, ok := rc.retryerMap.Load(m); ok {
				// NotifyPolicyChange may happen before
				continue
			}
			if err := rc.initRetryer(m, methodPolicies[m]); err != nil {
				rc.msg = err.Error()
				return err
			}
		}
	}
	rc.isCodeCfg = inited
	return nil
}

// DeletePolicy to delete the method by method.
func (rc *Container) DeletePolicy(method string) {
	rc.Lock()
	defer rc.Unlock()
	rc.msg = ""
	if rc.isCodeCfg {
		// the priority of user setup code policy is higher than remote config
		return
	}
	_, ok := rc.retryerMap.Load(method)
	if ok {
		rc.retryerMap.Delete(method)
		rc.msg = fmt.Sprintf("delete retryer[%s] at %s", method, time.Now())
	}
}

// NotifyPolicyChange to receive policy when it changes
func (rc *Container) NotifyPolicyChange(method string, p Policy) {
	rc.Lock()
	defer rc.Unlock()
	rc.msg = ""
	if rc.isCodeCfg {
		// the priority of user setup code policy is higher than remote config
		return
	}
	r, ok := rc.retryerMap.Load(method)
	if ok && r != nil {
		retryer, ok := r.(Retryer)
		if ok {
			if retryer.Type() == p.Type {
				retryer.UpdatePolicy(p)
				rc.msg = fmt.Sprintf("update retryer[%s-%s] at %s", method, retryer.Type(), time.Now())
				return
			}
			rc.retryerMap.Delete(method)
			rc.msg = fmt.Sprintf("delete retryer[%s-%s] at %s", method, retryer.Type(), time.Now())
		}
	}
	rc.initRetryer(method, p)
}

// PrepareRetryContext adds necessary keys to context for retry
// These keys should be added to `ctx` no matter whether there's a need to retry, to avoid sharing the same
// object objects with another method call, since `ctx` might be reused in user-defined middlewares.
func PrepareRetryContext(ctx context.Context) context.Context {
	// reqOp can be used to avoid multiple writes to the request object.
	// If a blocking write is needed, implement a lock based on it (spin-lock for example).
	reqOp := OpNo
	ctx = context.WithValue(ctx, ContextRequestOp, &reqOp)

	// `respOp` is used to avoid concurrent write/read on the response object, especially for backup requests.
	// If `respOp` is modified by one request of this method call, all other requests will skip decoding.
	respOp := OpNo
	ctx = context.WithValue(ctx, ContextRespOp, &respOp)
	return ctx
}
