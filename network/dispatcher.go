package network

func NewDefaultDispatcher(throughput int) dispatcher {
	return goroutineDispatcher(throughput)
}

type dispatcher interface {
	Schedule(fn func())
	Throughput() int
}

type goroutineDispatcher int

var _ dispatcher = goroutineDispatcher(0)

func (goroutineDispatcher) Schedule(fn func()) {
	go fn()
}

func (d goroutineDispatcher) Throughput() int {
	return int(d)
}
