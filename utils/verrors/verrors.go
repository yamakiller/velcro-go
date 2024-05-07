package verrors

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// 基础错误信息
var (
	ErrInternalException  = &basicError{"internal exception"}
	ErrServiceDiscovery   = &basicError{"service discovery error"}
	ErrLoadbalance        = &basicError{"loadbalance error"}
	ErrGetConnection      = &basicError{"get connection error"}
	ErrNoMoreInstance     = &basicError{"no more instances to retry"}
	ErrCircuitBreak       = &basicError{"forbidden by circuitbreaker"}
	ErrCanceledByBusiness = &basicError{"canceled by business"}
	ErrRPCTimeout         = &basicError{"rpc timeout"}
	ErrACL                = &basicError{"request forbidden"}
	ErrOverlimit          = &basicError{"request over limit"}
	ErrPanic              = &basicError{"panic"}

	ErrRetry = &basicError{"retry error"}
	// ErrRPCFinish 启用重试并且有一个呼叫已完成时发生
	ErrRPCFinish = &basicError{"rpc call finished"}
	// ErrRoute 当路由器无法路由此呼叫时发生
	ErrRoute = &basicError{"rpc route failed"}
)

var (
	ErrNotSupported  = ErrInternalException.WithCause(errors.New("operation not supported"))
	ErrNoDestService = ErrInternalException.WithCause(errors.New("no dest service"))
	ErrNoDestAddress = ErrInternalException.WithCause(errors.New("no dest address"))
	ErrNoConnection  = ErrInternalException.WithCause(errors.New("no connection available"))
	ErrConnOverLimit = ErrOverlimit.WithCause(errors.New("to many connections"))
	ErrQPSOverLimit  = ErrOverlimit.WithCause(errors.New("request too frequent"))

	ErrNoIvkRequest         = ErrInternalException.WithCause(errors.New("invoker request not set"))
	ErrServiceCircuitBreak  = ErrCircuitBreak.WithCause(errors.New("service circuitbreak"))
	ErrInstanceCircuitBreak = ErrCircuitBreak.WithCause(errors.New("instance circuitbreak"))
	ErrNoInstance           = ErrServiceDiscovery.WithCause(errors.New("no instance available"))
)

type basicError struct {
	message string
}

// Error 实现错误信息接口
func (be *basicError) Error() string {
	return be.message
}

// WithCause 创建一个详细的错误,将给定原因附加到当前错误.
func (be *basicError) WithCause(cause error) error {
	return &DetailedError{basic: be, cause: cause}
}

// WithCauseAndStack 创建一个详细的错误, 将给定的原因附加到当前错误和包装堆栈.
func (be *basicError) WithCauseAndStack(cause error, stack string) error {
	return &DetailedError{basic: be, cause: cause, stack: stack}
}

// WithCauseAndExtraMsg 创建一个详细的错误, 它将给定的原因附加到当前错误并包装额外的消息以提供错误消息.
func (be *basicError) WithCauseAndExtraMsg(cause error, extraMsg string) error {
	return &DetailedError{basic: be, cause: cause, extraMsg: extraMsg}
}

// Timeout 是否是超时错误
func (be *basicError) Timeout() bool {
	return be == ErrRPCTimeout
}

type DetailedError struct {
	basic    *basicError
	cause    error
	stack    string
	extraMsg string
}

func (de *DetailedError) Error() string {
	msg := appendErrMsg(de.basic.Error(), de.extraMsg)
	if de.cause != nil {
		return msg + ": " + de.cause.Error()
	}
	return msg
}

// Format the error.
func (de *DetailedError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			msg := appendErrMsg(de.basic.Error(), de.extraMsg)
			_, _ = io.WriteString(s, msg)
			if de.cause != nil {
				_, _ = fmt.Fprintf(s, ": %+v", de.cause)
			}
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, de.Error())
	}
}

// ErrorType returns the basic error type.
func (de *DetailedError) ErrorType() error {
	return de.basic
}

// Unwrap returns the cause of detailed error.
func (de *DetailedError) Unwrap() error {
	return de.cause
}

// Is returns if the given error matches the current error.
func (de *DetailedError) Is(target error) bool {
	return de == target || de.basic == target || errors.Is(de.cause, target)
}

// As returns if the given target matches the current error, if so sets
// target to the error value and returns true
func (de *DetailedError) As(target interface{}) bool {
	if errors.As(de.basic, target) {
		return true
	}
	return errors.As(de.cause, target)
}

// Timeout supports the os.IsTimeout checking.
func (de *DetailedError) Timeout() bool {
	return de.basic == ErrRPCTimeout || os.IsTimeout(de.cause)
}

// Stack record stack info
func (de *DetailedError) Stack() string {
	return de.stack
}

// WithExtraMsg to add extra msg to supply error msg
func (de *DetailedError) WithExtraMsg(extraMsg string) {
	de.extraMsg = extraMsg
}

func appendErrMsg(errMsg, extra string) string {
	if extra == "" {
		return errMsg
	}
	var strBuilder strings.Builder
	strBuilder.Grow(len(errMsg) + len(extra) + 2)
	strBuilder.WriteString(errMsg)
	strBuilder.WriteByte('[')
	strBuilder.WriteString(extra)
	strBuilder.WriteByte(']')
	return strBuilder.String()
}

// IsVelcroError 报告给定的错误是否是 velcro 生成的错误
func IsVelcroError(err error) bool {
	if _, ok := err.(*basicError); ok {
		return true
	}

	if _, ok := err.(*DetailedError); ok {
		return true
	}
	return false
}

// TimeoutCheckFunc 用于检查给定的err是否是超时错误.
var TimeoutCheckFunc func(err error) bool

// IsTimeoutError check if the error is timeout
func IsTimeoutError(err error) bool {
	if TimeoutCheckFunc != nil {
		return TimeoutCheckFunc(err)
	}
	return errors.Is(err, ErrRPCTimeout)
}
