package rpcinfo

import (
	"context"
	"net"
	"time"

	"github.com/yamakiller/velcro-go/rpc/transport"
	"github.com/yamakiller/velcro-go/rpc2/pkg/serviceinfo"
	"github.com/yamakiller/velcro-go/rpc2/pkg/stats"
	"github.com/yamakiller/velcro-go/utils/verrors"
)

type EndpointInfo interface {
	ServiceName() string
	Method() string
	Address() net.Addr
	Tag(key string) (value string, exist bool)
	DefaultTag(key, def string) string
}

// RPCStats is used to collect statistics about the RPC.
type RPCStats interface {
	Record(ctx context.Context, event stats.Event, status stats.Status, info string)
	SendSize() uint64
	RecvSize() uint64
	Error() error
	Panicked() (bool, interface{})
	GetEvent(event stats.Event) Event
	Level() stats.Level
	CopyForRetry() RPCStats
}

// Event is the abstraction of an event happened at a specific time.
type Event interface {
	Event() stats.Event
	Status() stats.Status
	Info() string
	Time() time.Time
	IsNil() bool
}

// Timeouts contains settings of timeouts.
type Timeouts interface {
	RPCTimeout() time.Duration
	ConnectTimeout() time.Duration
	ReadWriteTimeout() time.Duration
}

// TimeoutProvider provides timeout settings.
type TimeoutProvider interface {
	Timeouts(ri RPCInfo) Timeouts
}

// RPCConfig contains configuration for RPC.
type RPCConfig interface {
	Timeouts
	IOBufferSize() int
	TransportProtocol() transport.Protocol
	InteractionMode() InteractionMode
	PayloadCodec() serviceinfo.PayloadCodec
}

// Invocation contains specific information about the call.
type Invocation interface {
	PackageName() string
	ServiceName() string
	MethodName() string
	SeqID() int32
	BizStatusErr() verrors.BizStatusErrorIface
	Extra(key string) interface{}
}

// RPCInfo is the core abstraction of information about an RPC in Kitex.
type RPCInfo interface {
	From() EndpointInfo
	To() EndpointInfo
	Invocation() Invocation
	Config() RPCConfig
	Stats() RPCStats
}
