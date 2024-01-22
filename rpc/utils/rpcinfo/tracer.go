package rpcinfo

import (
	"context"
	"runtime/debug"

	"github.com/yamakiller/velcro-go/rpc/utils/stats"
	"github.com/yamakiller/velcro-go/vlog"
)

// TraceController controls tracers.
type TraceController struct {
	tracers []stats.Tracer
}

// Append appends a new tracer to the controller.
func (c *TraceController) Append(col stats.Tracer) {
	c.tracers = append(c.tracers, col)
}

// DoStart starts the tracers.
func (c *TraceController) DoStart(ctx context.Context, ri RPCInfo) context.Context {
	defer c.tryRecover(ctx)
	Record(ctx, ri, stats.RPCStart, nil)

	for _, col := range c.tracers {
		ctx = col.Start(ctx)
	}
	return ctx
}

// DoFinish calls the tracers in reversed order.
func (c *TraceController) DoFinish(ctx context.Context, ri RPCInfo, err error) {
	defer c.tryRecover(ctx)
	Record(ctx, ri, stats.RPCFinish, err)
	if err != nil {
		rpcStats := AsMutableRPCStats(ri.Stats())
		rpcStats.SetError(err)
	}

	// reverse the order
	for i := len(c.tracers) - 1; i >= 0; i-- {
		c.tracers[i].Finish(ctx)
	}
}

// HasTracer reports whether there exists any tracer.
func (c *TraceController) HasTracer() bool {
	return c != nil && len(c.tracers) > 0
}

func (c *TraceController) tryRecover(ctx context.Context) {
	if err := recover(); err != nil {
		vlog.ContextWarnf(ctx, "Panic happened during tracer call. This doesn't affect the rpc call, but may lead to lack of monitor data such as metrics and logs: error=%s, stack=%s", err, string(debug.Stack()))
	}
}
