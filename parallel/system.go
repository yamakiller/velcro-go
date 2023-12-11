package parallel

import (
	"net"
	"strconv"

	"github.com/yamakiller/velcro-go/eventstream"
	"github.com/yamakiller/velcro-go/logs"
)

type ParallelSystem struct {
	_handlers    *HandlerCollective
	_eventStream *eventstream.EventStream
	_deadLetter  *deathLetterHandler
	_opts        *OptionValue
	_stopper     chan struct{}
	_logger      logs.LogAgent
}

func (bx *ParallelSystem) Logger() logs.LogAgent {
	return bx._logger
}

//func (bx *ParallelSystem) NewLocalPID(id string) *PID {
//	return NewPID(bx._handlers._address, id)
//}

func (bx *ParallelSystem) SpawnNamed()

func (bx *ParallelSystem) Address() string {
	return bx._handlers._address
}

func (bx *ParallelSystem) GetHostPort() (host string, port int, err error) {
	addr := bx._handlers._address
	if h, p, e := net.SplitHostPort(addr); e != nil {
		if addr != localAddress {
			err = e
		}

		host = localAddress
		port = -1
	} else {
		host = h
		port, err = strconv.Atoi(p)
	}

	return
}

func (bx *ParallelSystem) Shutdown() {
	close(bx._stopper)
}

func (bx *ParallelSystem) IsStopped() bool {
	select {
	case <-bx._stopper:
		return true
	default:
		return false
	}
}
