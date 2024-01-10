package connectPool

import "net"

// DummyReporter is a dummy implementation of Reporter.
type DummyReporter struct{}

// ConnSucceed implements the Reporter interface.
func (dcm *DummyReporter) ConnSucceed(serviceName string, addr net.Addr) {
}

// ConnFailed implements the Reporter interface.
func (dcm *DummyReporter) ConnFailed(serviceName string, addr net.Addr) {
}

// ReuseSucceed implements the Reporter interface.
func (dcm *DummyReporter) ReuseSucceed(serviceName string, addr net.Addr) {
}
