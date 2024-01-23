package slt

import (
	"time"

	"github.com/yamakiller/velcro-go/rpc/utils/rpcinfo"
	"github.com/yamakiller/velcro-go/rpc/utils/stats"
)

// CalculateEventCost get events from rpcstats, and calculates the time duration of (end - start).
// It returns 0 when get nil rpcinfo event of either stats event.
func CalculateEventCost(rpcstats rpcinfo.RPCStats, start, end stats.Event) time.Duration {
	endEvent := rpcstats.GetEvent(end)
	startEvent := rpcstats.GetEvent(start)
	if startEvent == nil || endEvent == nil {
		return 0
	}
	return endEvent.Time().Sub(startEvent.Time())
}
