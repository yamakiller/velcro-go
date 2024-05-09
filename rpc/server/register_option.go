package server

import (
	internal_server "github.com/yamakiller/velcro-go/rpc/internal/server"
)

// RegisterOption 是配置服务注册的唯一方法.
type RegisterOption = internal_server.RegisterOption

// RegisterOptions ...
type RegisterOptions = internal_server.RegisterOptions

func WithFallbackService() RegisterOption {
	return RegisterOption{F: func(o *RegisterOptions) {
		o.IsFallbackService = true
	}}
}
