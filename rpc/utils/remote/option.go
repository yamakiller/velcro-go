package remote

import (
	"net"
	"time"

	"github.com/yamakiller/velcro-go/profiler"
	"github.com/yamakiller/velcro-go/rpc/utils/endpoint"
	"github.com/yamakiller/velcro-go/rpc/utils/rpcinfo"
	"github.com/yamakiller/velcro-go/rpc/utils/stats/serviceinfo"
)

// Option is used to pack the inbound and outbound handlers.
type Option struct {
	Outbounds []OutboundHandler

	Inbounds []InboundHandler

	StreamingMetaHandlers []StreamingMetaHandler
}

// PrependBoundHandler adds a BoundHandler to the head.
func (o *Option) PrependBoundHandler(h BoundHandler) {
	switch v := h.(type) {
	case DuplexBoundHandler:
		o.Inbounds = append([]InboundHandler{v}, o.Inbounds...)
		o.Outbounds = append([]OutboundHandler{v}, o.Outbounds...)
	case InboundHandler:
		o.Inbounds = append([]InboundHandler{v}, o.Inbounds...)
	case OutboundHandler:
		o.Outbounds = append([]OutboundHandler{v}, o.Outbounds...)
	default:
		panic("invalid BoundHandler: must implement InboundHandler or OutboundHandler")
	}
}

// AppendBoundHandler adds a BoundHandler to the end.
func (o *Option) AppendBoundHandler(h BoundHandler) {
	switch v := h.(type) {
	case DuplexBoundHandler:
		o.Inbounds = append(o.Inbounds, v)
		o.Outbounds = append(o.Outbounds, v)
	case InboundHandler:
		o.Inbounds = append(o.Inbounds, v)
	case OutboundHandler:
		o.Outbounds = append(o.Outbounds, v)
	default:
		panic("invalid BoundHandler: must implement InboundHandler or OutboundHandler")
	}
}

// ServerOption contains option that is used to init the remote server.
type ServerOption struct {
	SvcMap map[string]*serviceinfo.ServiceInfo

	TransServerFactory TransServerFactory

	SvrHandlerFactory ServerTransHandlerFactory

	Codec Codec

	PayloadCodec PayloadCodec

	// Listener is used to specify the server listener, which comes with higher priority than Address below.
	Listener net.Listener

	// Address is the listener addr
	Address net.Addr

	ReusePort bool

	// Duration that server waits for to allow any existing connection to be closed gracefully.
	ExitWaitTime time.Duration

	// Duration that server waits for after error occurs during connection accepting.
	AcceptFailedDelayTime time.Duration

	// Duration that the accepted connection waits for to read or write data.
	MaxConnectionIdleTime time.Duration

	ReadWriteTimeout time.Duration

	InitOrResetRPCInfoFunc func(rpcinfo.RPCInfo, net.Addr) rpcinfo.RPCInfo

	TracerCtl *rpcinfo.TraceController

	Profiler                 profiler.Profiler
	ProfilerTransInfoTagging TransInfoTagging
	ProfilerMessageTagging   MessageTagging

	//GRPCCfg *grpc.ServerConfig

	//GRPCUnknownServiceHandler func(ctx context.Context, method string, stream streaming.Stream) error

	Option

	// invoking chain with recv/send middlewares for streaming APIs
	RecvEndpoint endpoint.RecvEndpoint
	SendEndpoint endpoint.SendEndpoint

	// for thrift streaming, this is enabled by default
	// for grpc(protobuf) streaming, it's disabled by default, enable with server.WithCompatibleMiddlewareForUnary
	CompatibleMiddlewareForUnary bool
}

// ClientOption is used to init the remote client.
type ClientOption struct {
	SvcInfo *serviceinfo.ServiceInfo

	CliHandlerFactory ClientTransHandlerFactory

	Codec Codec

	PayloadCodec PayloadCodec

	ConnPool ConnPool

	Dialer Dialer

	Option

	EnableConnPoolReporter bool
}
