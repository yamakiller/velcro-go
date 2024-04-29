package circuitbreak

import (
	"context"
	"errors"
	"time"

	"github.com/yamakiller/velcro-go/rpc/utils/verrors"
	"github.com/yamakiller/velcro-go/utils/endpoint"
)

// State changes between Closed, Open, HalfOpen
// [Closed] -->- tripped ----> [Open]<-------+
//
//	^                          |           ^
//	|                          v           |
//	+                          |      detect fail
//	|                          |           |
//	|                    cooling timeout   |
//	^                          |           ^
//	|                          v           |
//	+--- detect succeed --<-[HalfOpen]-->--+
//
// The behaviors of each states:
// =================================================================================================
// |           | [Succeed]                  | [Fail or Timeout]       | [IsAllowed]                |
// |================================================================================================
// | [Closed]  | do nothing                 | if tripped, become Open | allow                      |
// |================================================================================================
// | [Open]    | do nothing                 | do nothing              | if cooling timeout, allow; |
// |           |                            |                         | else reject                |
// |================================================================================================
// |           |increase halfopenSuccess,   |                         | if detect timeout, allow;  |
// |[HalfOpen] |if(halfopenSuccess >=       | become Open             | else reject                |
// |           | defaultHalfOpenSuccesses)|                         |                            |
// |           |     become Closed          |                         |                            |
// =================================================================================================
type State int32

func (s State) String() string {
	switch s {
	case Open:
		return "OPEN"
	case HalfOpen:
		return "HALFOPEN"
	case Closed:
		return "CLOSED"
	}
	return "INVALID"
}

// represents the state
const (
	Open     State = iota
	HalfOpen State = iota
	Closed   State = iota
)

// BreakerStateChangeHandler .
type BreakerStateChangeHandler func(oldState, newState State, m Metricer)

// PanelStateChangeHandler .
type PanelStateChangeHandler func(key string, oldState, newState State, m Metricer)

// Options for breaker
type Options struct {
	// parameters for metricser
	BucketTime time.Duration // the time each bucket holds
	BucketNums int32         // the number of buckets the breaker have

	// parameters for breaker
	CoolingTimeout    time.Duration // fixed when create
	DetectTimeout     time.Duration // fixed when create
	HalfOpenSuccesses int32         // halfopen success is the threshold when the breaker is in HalfOpen;
	// after exceeding consecutively this times, it will change its State from HalfOpen to Closed;

	ShouldTrip                TripFunc                  // can be nil
	ShouldTripWithKey         TripFuncWithKey           // can be nil, overwrites ShouldTrip
	BreakerStateChangeHandler BreakerStateChangeHandler // can be nil

	// if to use Per-P Metricer
	// use Per-P Metricer can increase performance in multi-P condition
	// but will consume more memory
	EnableShardP bool

	// Default value is time.Now, caller may use some high-performance custom time now func here
	Now func() time.Time
}

const (
	// bucket time is the time each bucket holds
	defaultBucketTime = time.Millisecond * 100

	// bucket nums is the number of buckets the metricser has;
	// the more buckets you have, the less counters you lose when
	// the oldest bucket expire;
	defaultBucketNums = 100

	// default window size is (defaultBucketTime * defaultBucketNums),
	// which is 10 seconds;
)

// Parameter 包含断路器参数
type Parameter struct {
	// Enabled 是否启动断路器.
	Enabled bool
	// ErrorRate 表示断裂的速率.
	ErrorRate float64
	// MinimalSample 表示断裂前需要的最小采样.
	MinimalSample int64
}

// ErrorType 表示错误类型.
type ErrorType int

// Constants for ErrorType.
const (
	// TypeIgnorable 表示可忽略错误,被断路器忽略.
	TypeIgnorable ErrorType = iota
	// TypeTimeout 表示超时错误.
	TypeTimeout
	// TypeFailure 表示请求失败,但不是超时.
	TypeFailure
	// TypeSuccess 表示请求成功.
	TypeSuccess
)

// WrapErrorWithType 用于定义 CircuitBreaker 的 ErrorType;
// 如果不希望错误触发熔断, 可以将ErrorType设置为TypeIgnorable,错误不会被视为失败;
// eg: 在自定义中间件中返回 Circuitbreak.WrapErrorWithType.WithCause(err, Circuitbreak.TypeIgnorable).
func WrapErrorWithType(err error, errorType ErrorType) CircuitBreakerAwareError {
	return &errorWrapperWithType{err: err, errType: errorType}
}

type GetErrorTypeFunc func(ctx context.Context, request, response interface{}, err error) ErrorType

// Control is the control strategy of the circuit breaker.
type Control struct {
	// 实现这将为断路器面板生成密钥.
	GetKey func(ctx context.Context, request interface{}) (key string, enabled bool)

	// 实现这个以确定错误的类型.
	GetErrorType GetErrorTypeFunc

	// 实施此操作可提供有关断路器的更多详细信息.
	// err 参数始终是 verrors.ErrCircuitBreak.
	DecorateError func(ctx context.Context, request interface{}, err error) error
}

// NewCircuitBreakerMW  使用给定的控制策略和面板创建断路器 MW.
func NewCircuitBreakerMW(control Control, panel Panel) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request, response interface{}) (err error) {
			key, enabled := control.GetKey(ctx, request)
			if !enabled {
				return next(ctx, request, response)
			}

			if !panel.IsAllowed(key) {
				return control.DecorateError(ctx, request, verrors.ErrCircuitBreak)
			}

			err = next(ctx, request, response)
			RecordStat(ctx, request, response, err, key, &control, panel)
			return
		}
	}
}

// RecordStat 向断路器报告请求结果
func RecordStat(ctx context.Context, request, response interface{}, err error, cbKey string, ctl *Control, panel Panel) {
	switch ctl.GetErrorType(ctx, request, response, err) {
	case TypeTimeout:
		panel.Timeout(cbKey)
	case TypeFailure:
		panel.Fail(cbKey)
	case TypeSuccess:
		panel.Succeed(cbKey)
	}
}

// CircuitBreakerAwareError 用于包装ErrorType
type CircuitBreakerAwareError interface {
	error
	TypeForCircuitBreaker() ErrorType
}

type errorWrapperWithType struct {
	errType ErrorType
	err     error
}

func (e errorWrapperWithType) TypeForCircuitBreaker() ErrorType {
	return e.errType
}

func (e errorWrapperWithType) Error() string {
	return e.err.Error()
}

func (e errorWrapperWithType) Unwrap() error {
	return e.err
}

func (e errorWrapperWithType) Is(target error) bool {
	return errors.Is(e.err, target)
}
