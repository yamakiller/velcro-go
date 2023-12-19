package p2c

import (
	"math"
	"time"
)

const (
	// Name is the name of p2c balancer.
	Name = "p2c_ewma"

	decayTime       = int64(time.Second * 10) // default value from finagle
	forcePick       = int64(time.Second)
	initSuccess     = 1000
	throttleSuccess = initSuccess / 2
	penalty         = int64(math.MaxInt32)
	pickTimes       = 3
	logInterval     = time.Minute
)

type p2cPicker struct {
	//conns []*subConn
	//r     *rand.Rand
	//stamp *syncx.AtomicDuration
	//lock  sync.Mutex
}
