package loadbalance

import (
	"sync/atomic"

	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
)

// round 显现一个 严格的循环算法
type round struct {
	state uint64    // 8 bytes
	_     [7]uint64 // + 7 * 8 bytes
	// = 64 bytes
}

func (r *round) Next() uint64 {
	return atomic.AddUint64(&r.state, 1)
}

func newRound() *round {
	r := &round{}
	return r
}

func newRandomRound() *round {
	r := &round{
		state: fastrand.Uint64(),
	}
	return r
}
