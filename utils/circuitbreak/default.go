package circuitbreak

import (
	"context"
	"errors"

	"github.com/yamakiller/velcro-go/utils/verrors"
)

// some types of error won't trigger circuit breaker
var ignoreErrTypes = map[error]ErrorType{
	verrors.ErrInternalException: TypeIgnorable,
	verrors.ErrServiceDiscovery:  TypeIgnorable,
	verrors.ErrACL:               TypeIgnorable,
	verrors.ErrLoadbalance:       TypeIgnorable,
	verrors.ErrRPCFinish:         TypeIgnorable,
}

// ErrorTypeOnServiceLevel 使用服务级别标准确定错误类型.
func ErrorTypeOnServiceLevel(ctx context.Context, request, response interface{}, err error) ErrorType {
	if err != nil {
		for e, t := range ignoreErrTypes {
			if errors.Is(err, e) {
				return t
			}
		}
		var we *errorWrapperWithType
		if ok := errors.As(err, &we); ok {
			return we.errType
		}
		if verrors.IsTimeoutError(err) {
			return TypeTimeout
		}
		return TypeFailure
	}
	return TypeSuccess
}

// ErrorTypeOnInstanceLevel 使用实例级别标准确定错误类型.
// 基本上，它仅将连接错误视为失败.
func ErrorTypeOnInstanceLevel(ctx context.Context, request, response interface{}, err error) ErrorType {
	if errors.Is(err, verrors.ErrGetConnection) {
		return TypeFailure
	}
	return TypeSuccess
}

// FailIfError return TypeFailure if err is not nil, otherwise TypeSuccess.
func FailIfError(ctx context.Context, request, response interface{}, err error) ErrorType {
	if err != nil {
		return TypeFailure
	}
	return TypeSuccess
}

// NoDecoration returns the original err.
func NoDecoration(ctx context.Context, request interface{}, err error) error {
	return err
}
