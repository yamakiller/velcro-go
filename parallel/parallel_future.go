package parallel

import (
	"errors"
	sync "sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	ErrorTimeout     = errors.New("future: timeout")
	ErrorDeathLetter = errors.New("future: death letter")
)

func NewFuture(system *ParallelSystem, d time.Duration) *Future {
	ref := &futureHandler{Future{_system: system, _cond: sync.NewCond(&sync.Mutex{})}}
	id := system._handlers.NextId()

	pid, ok := system._handlers.Push(ref, "future"+id)
	if !ok {
		system.Logger().Error("[FUTURE]", "failed to register future process [pid:%+v]", pid)
	}

	ref._pid = pid

	if d >= 0 {
		tp := time.AfterFunc(d, func() {
			ref._cond.L.Lock()
			if ref._done {
				ref._cond.L.Unlock()

				return
			}
			ref._err = ErrorTimeout
			ref._cond.L.Unlock()
			ref.Stop(pid)
		})
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&ref._t)), unsafe.Pointer(tp))
	}

	return &ref.Future
}

type Future struct {
	_system *ParallelSystem
	_pid    *PID
	_cond   *sync.Cond
	// protected by _cond
	_done        bool
	_result      interface{}
	_err         error
	_t           *time.Timer
	_pipes       []*PID
	_completions []func(res interface{}, err error)
}

func (f *Future) PID() *PID {
	return f._pid
}

// PipeTo forwards the _result or error of the future to the specified pids.
func (f *Future) PipeTo(pids ...*PID) {
	f._cond.L.Lock()
	f._pipes = append(f._pipes, pids...)
	// for an already completed future, force push the _result to targets.
	if f._done {
		f.sendToPipes()
	}
	f._cond.L.Unlock()
}

func (f *Future) sendToPipes() {
	if f._pipes == nil {
		return
	}

	var m interface{}
	if f._err != nil {
		m = f._err
	} else {
		m = f._result
	}

	for _, pid := range f._pipes {
		pid.postUsrMessage(f._system, m)
	}

	f._pipes = nil
}

func (f *Future) wait() {
	f._cond.L.Lock()
	for !f._done {
		f._cond.Wait()
	}
	f._cond.L.Unlock()
}

// Result waits for the future to resolve.
func (f *Future) Result() (interface{}, error) {
	f.wait()

	return f._result, f._err
}

func (f *Future) Wait() error {
	f.wait()

	return f._err
}

func (f *Future) continueWith(continuation func(res interface{}, err error)) {
	f._cond.L.Lock()
	defer f._cond.L.Unlock() // use defer as the continuation co
	// uld blow up
	if f._done {
		continuation(f._result, f._err)
	} else {
		f._completions = append(f._completions, continuation)
	}
}

// futureHandler is a struct carrying a response PID and a channel where the response is placed.
type futureHandler struct {
	Future
}

func (ref *futureHandler) postUsrMessage(pid *PID, message interface{}) {
	//defer ref.instrument()

	_, msg, _ := UnWrapReport(message)

	if _, ok := msg.(*DeadLetterResponse); ok {
		ref._result = nil
		ref._err = ErrorDeathLetter
	} else {
		ref._result = msg
	}

	ref.Stop(pid)
}

func (ref *futureHandler) postSysMessage(pid *PID, message interface{}) {
	//defer ref.instrument()
	ref._result = message
	ref.Stop(pid)
}

/*func (ref *futureProcess) instrument() {
	sysMetrics, ok := ref._system.Extensions.Get(extensionId).(*Metrics)
	if ok && sysMetrics.enabled {
		ctx := context.Background()
		labels := []attribute.KeyValue{
			attribute.String("address", ref.actorSystem.Address()),
		}

		instruments := sysMetrics.metrics.Get(metrics.InternalActorMetrics)
		if instruments != nil {
			if ref.err == nil {
				instruments.FuturesCompletedCount.Add(ctx, 1, metric.WithAttributes(labels...))
			} else {
				instruments.FuturesTimedOutCount.Add(ctx, 1, metric.WithAttributes(labels...))
			}
		}
	}
}*/

func (ref *futureHandler) Stop(pid *PID) {
	ref._cond.L.Lock()
	if ref._done {
		ref._cond.L.Unlock()

		return
	}

	ref._done = true
	tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&ref._t))))

	if tp != nil {
		tp.Stop()
	}

	ref._system._handlers.Erase(pid)

	ref.sendToPipes()
	ref.runCompletions()
	ref._cond.L.Unlock()
	ref._cond.Signal()
}

func (f *Future) overloadUsrMessage() int {
	return 0
}

func (f *Future) runCompletions() {
	if f._completions == nil {
		return
	}

	for _, c := range f._completions {
		c(f._result, f._err)
	}

	f._completions = nil
}
