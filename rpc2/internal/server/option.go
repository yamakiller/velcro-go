package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yamakiller/velcro-go/cluster/utils/limit"
	"github.com/yamakiller/velcro-go/cluster/utils/limiter"
	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/utils"
)

type Option struct {
	F func(o *Options, di *utils.Slice)
}

type Options struct {
	Limit Limit
	// 退出信号
	ExitSignal func() <-chan error

	DebugInfo utils.Slice
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
