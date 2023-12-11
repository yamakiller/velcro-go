package coroutine

import (
	"runtime"
	"sync"
	"unsafe"
)

var (
	defaultContext *context = nil
)

func init() {
	defaultContext = &context{
		_minOfNumber:      runtime.NumCPU() * 2,
		_processWaitGroup: sync.WaitGroup{},
		//_processIdle:      make([]*process, runtime.NumCPU()*2),
	}
}

type process struct {
	_id     int32
	_state  State
	_cond   sync.Cond
	_stoped bool
	_func   func(...interface{})
	_args   []interface{}

	_prev *process
	_next *process
}

func (p *process) post(f func(...interface{}), args ...interface{}) {
	for _, arg := range args {
		p._args = append(p._args, arg)
	}
	p._func = f
	p._cond.Signal()
}

func (p *process) run(ctx *context) {
	defer ctx._processWaitGroup.Done()
	for {

		if p._stoped {
			break
		}

		p._cond.L.Lock()
		for !p._stoped {
			p._cond.Wait()
		}
		p._cond.L.Unlock()

		if p._func != nil {
			p._state = RunState
			p._func(p._args...)
			p._func = nil
			if len(p._args) != 0 {
				p._args = make([]interface{}, 0)
			}
			p._state = IdleState
			ctx.pushIdle(p)
		}
	}

	p._state = DeadState
	ctx.eraseIdle(p)
}

type context struct {
	_seq              int32
	_minOfNumber      int
	_processWaitGroup sync.WaitGroup
	_process          map[unsafe.Pointer]*process
	_idleHead         *process
	_idleTail         *process
	_processSync      sync.Mutex
}

func (ctx *context) malloc() *process {
	var pc *process = nil
	defaultContext._processSync.Lock()
	pc = ctx.popIdle()

	defaultContext._processSync.Unlock()

	if pc == nil {
		pc = &process{_state: IdleState, _prev: nil, _next: nil}
		defaultContext._processSync.Lock()
		defaultContext._process[unsafe.Pointer(pc)] = pc
		defaultContext._processSync.Unlock()
		defaultContext._processWaitGroup.Add(1)
		go pc.run(defaultContext)
	}

	return pc
}

func (ctx *context) popIdle() *process {
	if ctx._idleHead == nil {
		return nil
	}

	var p *process = ctx._idleHead

	ctx._idleHead = ctx._idleHead._next
	if ctx._idleHead == nil {
		ctx._idleTail = nil
	} else {
		ctx._idleHead._prev = nil
	}

	return p
}

func (ctx *context) pushIdle(p *process) {
	ctx._processSync.Lock()
	defer ctx._processSync.Unlock()

	if ctx._idleTail == nil {
		ctx._idleHead = p
	} else {
		p._prev = ctx._idleTail
	}

	ctx._idleTail = p
}

func (ctx *context) eraseIdle(p *process) {
	ctx._processSync.Lock()
	defer ctx._processSync.Unlock()

	if p._prev != nil {
		p._prev._next = p._next
	} else {
		ctx._idleHead = p._next
	}

	if p._next != nil {
		p._next._prev = p._prev
	} else {
		ctx._idleTail = p._prev
	}
}

func Go(f func(...interface{}), args ...interface{}) {
	defaultContext.malloc().post(f, args...)
}

func WaitShutdown() {

	defaultContext._processWaitGroup.Wait()
}
