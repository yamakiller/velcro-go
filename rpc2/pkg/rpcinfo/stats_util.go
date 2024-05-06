package rpcinfo

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/yamakiller/velcro-go/rpc2/pkg/stats"
	"github.com/yamakiller/velcro-go/vlog"
)

// Record records the event to RPCStats.
func Record(ctx context.Context, ri RPCInfo, event stats.Event, err error) {
	if ctx == nil || ri.Stats() == nil {
		return
	}
	if err != nil {
		ri.Stats().Record(ctx, event, stats.StatusError, err.Error())
	} else {
		ri.Stats().Record(ctx, event, stats.StatusInfo, "")
	}
}

// CalcEventCostUs calculates the duration between start and end and returns in microsecond.
func CalcEventCostUs(start, end Event) uint64 {
	if start == nil || end == nil || start.IsNil() || end.IsNil() {
		return 0
	}
	return uint64(end.Time().Sub(start.Time()).Microseconds())
}

// ClientPanicToErr to transform the panic info to error, and output the error if needed.
func ClientPanicToErr(ctx context.Context, panicInfo interface{}, ri RPCInfo, logErr bool) error {
	e := fmt.Errorf("KITEX: client panic, to_service=%s to_method=%s error=%v\nstack=%s",
		ri.To().ServiceName(), ri.To().Method(), panicInfo, debug.Stack())
	rpcStats := AsMutableRPCStats(ri.Stats())
	rpcStats.SetPanicked(e)
	if logErr {
		vlog.ContextErrorf(ctx, "%s", e.Error())
	}
	return e
}
