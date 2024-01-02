package clientpool

import (
	"net"
	"sync"
)

var (
	isSet          bool
	rwLock         sync.RWMutex
	commonReporter Reporter = &DummyReporter{}
)

type Reporter interface {
	ConnSucceed(serviceName string, addr net.Addr)
	ConnFailed(serviceName string, addr net.Addr)
	ReuseSucceed(serviceName string, addr net.Addr)
}

// SetReporter set the common reporter of connection pool, that can only be set once.
func SetReporter(r Reporter) {
	rwLock.Lock()
	defer rwLock.Unlock()
	if !isSet {
		isSet = true
		commonReporter = r
	}
}

// GetCommonReporter returns the current Reporter used by connection pools.
func GetCommonReporter() Reporter {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return commonReporter
}
