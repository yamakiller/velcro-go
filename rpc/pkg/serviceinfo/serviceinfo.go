package serviceinfo

import "context"

// PayloadCodec 有效数据编码器类型别名
type PayloadCodec int

const (
	Thrift PayloadCodec = iota
	Protobuf
	Hessian2
)

const (
	// GenericService name 服务名字
	GenericService = "$GenericService" // private as "$"
	// GenericMethod name
	GenericMethod = "$GenericCall"
)

// ServiceInfo 记录服务的元信息
type ServiceInfo struct {
	// 服务的名称. 对于通用服务, 它始终是常量"GenericService";
	ServiceName string

	// HandlerType 是生成代码中请求处理程序的类型值;
	HandlerType interface{}

	// Methods 包含服务支持的方法的元信息;
	// 对于通用服务, 只有一个由常量"GenericMethod"命名的方法.
	Methods map[string]MethodInfo

	// PayloadCodec 数据编码器
	PayloadCodec PayloadCodec

	// VelcroGenVersion 版本信息,命令行工具 'velcroCodec'
	VelcroGenVersion string

	// Extra 用于未来的功能信息, 以避免兼容性问题;
	// 否则我们需要在结构中添加一个新字段.
	Extra map[string]interface{}

	// GenericMethod 返回给定名称的 MethodInfo
	// 它仅由通用调用使用.
	GenericMethod func(name string) MethodInfo
}

func (i *ServiceInfo) MethodInfo(name string) MethodInfo {
	if i == nil {
		return nil
	}
	if i.ServiceName == GenericService {
		if i.GenericMethod != nil {
			return i.GenericMethod(name)
		}
		return i.Methods[name]
	}
	return i.Methods[name]
}

type StreamingMode int

const (
	StreamingNode          StreamingMode = 0b0000 // VelcroPB/VelcroThrift
	StreamingUnary         StreamingMode = 0b0001 // 流式底层传输协议, 如: HTTP2
	StreamingClient        StreamingMode = 0b0010
	StreamingServer        StreamingMode = 0b0100
	StreamingBidirectional StreamingMode = 0b0110
)

type MethodInfo interface {
	Handler() MethodHandler
	NewArgs() interface{}
	NewResult() interface{}
	OneWay() bool
	IsStreaming() bool
	StreamingMode() StreamingMode
}

// MethodHandler 对应于生成代码中的处理程序包装函数.
type MethodHandler func(ctx context.Context, handler, args, result interface{}) error

// MethodInfoOption 修改给定的 MethodInfo.
type MethodInfoOption func(*methodInfo)

// WithStreamingMode 设置流模式并相应更新流标志.
func WithStreamingMode(mode StreamingMode) MethodInfoOption {
	return func(m *methodInfo) {
		m.streamingMode = mode
		m.isStreaming = mode != StreamingNode
	}
}

func NewMethodInfo(methodHandler MethodHandler, newArgsFunc, newResultFunc func() interface{}, oneWay bool, opts ...MethodInfoOption) MethodInfo {
	mi := methodInfo{
		handler:       methodHandler,
		newArgsFunc:   newArgsFunc,
		newResultFunc: newResultFunc,
		oneWay:        oneWay,
		isStreaming:   false,
		streamingMode: StreamingNode,
	}

	for _, opt := range opts {
		opt(&mi)
	}

	return &mi
}

type methodInfo struct {
	handler       MethodHandler
	newArgsFunc   func() interface{}
	newResultFunc func() interface{}
	oneWay        bool
	isStreaming   bool
	streamingMode StreamingMode
}

// Handler 实现 MethodInfo 接口.
func (m methodInfo) Handler() MethodHandler {
	return m.handler
}

// NewArgs 实现 MethodInfo 接口.
func (m methodInfo) NewArgs() interface{} {
	return m.newArgsFunc()
}

// NewResult 实现 MethodInfo 接口.
func (m methodInfo) NewResult() interface{} {
	return m.newResultFunc()
}

// OneWay 实现 MethodInfo 接口.
func (m methodInfo) OneWay() bool {
	return m.oneWay
}

func (m methodInfo) IsStreaming() bool {
	return m.isStreaming
}

func (m methodInfo) StreamingMode() StreamingMode {
	return m.streamingMode
}

func (p PayloadCodec) String() string {
	switch p {
	case Thrift:
		return "Thrift"
	case Protobuf:
		return "Protobuf"
	case Hessian2:
		return "Hessian2"
	default:
		panic("unknown payload type")
	}
}
