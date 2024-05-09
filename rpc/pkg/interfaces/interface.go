package interfaces

// VelcroArgs 当从 ?Args 获取真实请求时用于断言.
// Thrift and VelcroProtobuf 将生成 VelcroArgs 接口 => GetFirstArgument(){}
type VelcroArgs interface {
	GetFirstArgument() interface{}
}

// VelcroResult 当从 ?Result 得到真实响应时用于断言.
// Thrift 和 VelcroProtobuf 将为 VelcroResult 生成两个函数
type VelcroResult interface {
	GetResult() interface{}
	SetSuccess(interface{})
}
