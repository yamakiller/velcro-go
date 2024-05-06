package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yamakiller/velcro-go/cluster/utils/limit"
	"github.com/yamakiller/velcro-go/cluster/utils/limiter"
	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/rpc2/pkg/registry"
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote/codec/thrift"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/rpc2/pkg/serviceinfo"
	"github.com/yamakiller/velcro-go/rpc2/pkg/stats"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/acl"
	"github.com/yamakiller/velcro-go/utils/configutil"
	"github.com/yamakiller/velcro-go/utils/endpoint"
)

func init() {
	remote.PutPayloadCode(serviceinfo.Thrift, thrift.NewThriftCodec())
	//remote.PutPayloadCode(serviceinfo.Protobuf, protobuf.NewProtobufCodec())
}

func NewOptions(opts []Option) *Options {
	o := &Options{
		Configs:    rpcinfo.NewRPCConfig(),
		Once:       configutil.NewOptionOnce(),
		RemoteOpt:  newServerRemoteOption(),
		ExitSignal: DefaultSysExitSignal,
		Registry:   registry.NoopRegistry,
	}
	ApplyOptions(opts, o)
	rpcinfo.AsMutableRPCConfig(o.Configs).LockConfig(o.LockBits)
	if o.StatsLevel == nil {
		level := stats.LevelDisabled
		if o.TracerCtl.HasTracer() {
			level = stats.LevelDetailed
		}
		o.StatsLevel = &level
	}
	return o
}

type Option struct {
	F func(o *Options, di *utils.Slice)
}

type Options struct {
	Configs  rpcinfo.RPCConfig
	LockBits int
	Once     *configutil.OptionOnce

	RemoteOpt *remote.ServerOption
	ErrHandle func(context.Context, error) error
	// 退出信号
	ExitSignal func() <-chan error

	Registry     registry.Registry
	RegistryInfo *registry.Info

	ACLRules []acl.RejectFunc
	Limit    Limit
	MWBs     []endpoint.RecvMiddlewareBuilder

	DebugInfo utils.Slice

	TracerCtl  *rpcinfo.TraceController
	StatsLevel *stats.Level

	RefuseTrafficWithoutServiceName bool
	EnableContextTimeout            bool
}

type Limit struct {
	Limits        *limit.Option
	LimitReporter limiter.LimitReporter
	ConnLimit     limiter.ConcurrencyLimiter
	QPSLimit      limiter.RateLimiter

	QPSLimitPostDecode bool
}

// ApplyOptions applies the given options.
func ApplyOptions(opts []Option, o *Options) {
	for _, op := range opts {
		op.F(o, &o.DebugInfo)
	}
}

func DefaultSysExitSignal() <-chan error {
	errCh := make(chan error, 1)
	gofunc.GoFunc(context.Background(), func() {
		sig := SysExitSignal()
		defer signal.Stop(sig)
		<-sig
		errCh <- nil
	})
	return errCh
}

func SysExitSignal() chan os.Signal {
	signals := make(chan os.Signal, 1)
	notifications := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	if !signal.Ignored(syscall.SIGHUP) {
		notifications = append(notifications, syscall.SIGHUP)
	}
	signal.Notify(signals, notifications...)
	return signals
}
