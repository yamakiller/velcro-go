package server

// RegisterOption 是配置服务注册的唯一方法
type RegisterOption struct {
	F func(o *RegisterOptions)
}

// RegisterOptions 用于配置服务注册
type RegisterOptions struct {
	IsFallbackService bool
}

func NewREgisterOptions(opts []RegisterOption) *RegisterOptions {
	o := &RegisterOptions{}
	ApplyRegisterOptions(opts, o)
	return o
}

// ApplyRegisterOptions applies the given register options.
func ApplyRegisterOptions(opts []RegisterOption, o *RegisterOptions) {
	for _, op := range opts {
		op.F(o)
	}
}
