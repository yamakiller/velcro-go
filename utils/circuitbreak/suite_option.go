package circuitbreak

import "context"

type SuiteConfig struct {
	serviceGetErrorTypeFunc  GetErrorTypeFunc
	instanceGetErrorTypeFunc GetErrorTypeFunc
}

type SuiteOption func(s *SuiteConfig)

// WithServiceGetErrorType sets serviceControl.GetErrorType
// Velcro 将调用customFunc来确定断路器的错误类型,建议用户使用
// WithWrappedServiceGetErrorType来保留大部分行为.
// 注意: 这用于服务级断路器
func WithServiceGetErrorType(customFunc GetErrorTypeFunc) SuiteOption {
	return func(cfg *SuiteConfig) {
		cfg.serviceGetErrorTypeFunc = customFunc
	}
}

// WithWrappedServiceGetErrorType sets serviceControl.GetErrorType
// Velcro 首先会调用ErrorTypeOnServiceLevel,如果返回TypeSuccess, 则会调用
// customFunc来确定断路器的最终错误类型.
func WithWrappedServiceGetErrorType(customFunc GetErrorTypeFunc) SuiteOption {
	return func(cfg *SuiteConfig) {
		cfg.serviceGetErrorTypeFunc = WrapErrorTypeFunc(customFunc, ErrorTypeOnServiceLevel)
	}
}

// WithInstanceGetErrorType sets instanceControl.GetErrorType
// 会调用customFunc来确定断路器的错误类型建议用户使用WithWrappedInstanceGetErrorType
// 来保留大部分行为.
// 注意: 这用于实例级断路器
func WithInstanceGetErrorType(f GetErrorTypeFunc) SuiteOption {
	return func(cfg *SuiteConfig) {
		cfg.instanceGetErrorTypeFunc = f
	}
}

// WithWrappedInstanceGetErrorType sets instanceControl.GetErrorType
// 首先会调用ErrorTypeOnInstanceLevel, 如果返回TypeSuccess,则会调用
// customFunc来确定断路器的最终错误类型.
// 注意: 这用于实例级断路器
func WithWrappedInstanceGetErrorType(f GetErrorTypeFunc) SuiteOption {
	return func(cfg *SuiteConfig) {
		cfg.instanceGetErrorTypeFunc = WrapErrorTypeFunc(f, ErrorTypeOnInstanceLevel)
	}
}

// WrapErrorTypeFunc 如果originalFunc返回TypeSuccess, 则调用customFunc.
// customFunc可以根据业务需求选择性地返回其他类型
func WrapErrorTypeFunc(customFunc, originalFunc GetErrorTypeFunc) GetErrorTypeFunc {
	return func(ctx context.Context, request, response interface{}, err error) ErrorType {
		if errorType := originalFunc(ctx, request, response, err); errorType != TypeSuccess {
			return errorType
		}
		return customFunc(ctx, request, response, err)
	}
}
