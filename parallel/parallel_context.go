package parallel

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/logs"
)

const (
	stateAlive int32 = iota
	stateRestarting
	stateStopping
	stateStopped
)

type Context interface {
	// Self returns the PID for the current parallel
	Self() *PID

	Info(sfmt string, args ...interface{})

	Debug(sfmt string, args ...interface{})

	Error(sfmt string, args ...interface{})

	Warning(sfmt string, args ...interface{})

	Fatal(sfmt string, args ...interface{})

	Panic(sfmt string, args ...interface{})
}

func newParallelContext(system *ParallelSystem) *parallelContext {

	this := &parallelContext{
		_system: system,
	}

	this.instantiateParallel()

	return this
}

type parallelContextExtras struct {
	_receiveTimeoutTimer *time.Timer
	_rs                  *RestartStatistics
	_watchers            PIDSet
}

func (pcextr *parallelContextExtras) restartStats() *RestartStatistics {

	// 创建重启统计信息
	if pcextr._rs == nil {
		pcextr._rs = NewRestartStatistics()
	}

	return pcextr._rs
}

func (pcextr *parallelContextExtras) initReceiveTimeoutTimer(timer *time.Timer) {
	pcextr._receiveTimeoutTimer = timer
}

func (pcextr *parallelContextExtras) resetReceiveTimeoutTimer(time time.Duration) {
	if pcextr._receiveTimeoutTimer == nil {
		return
	}

	pcextr._receiveTimeoutTimer.Reset(time)
}

func (pcextr *parallelContextExtras) stopReceiveTimeoutTimer() {
	if pcextr._receiveTimeoutTimer == nil {
		return
	}

	pcextr._receiveTimeoutTimer.Stop()
}

func (pcextr *parallelContextExtras) killReceiveTimeoutTimer() {
	if pcextr._receiveTimeoutTimer == nil {
		return
	}

	pcextr._receiveTimeoutTimer.Stop()
	pcextr._receiveTimeoutTimer = nil
}

func (pcextr *parallelContextExtras) watch(watcher *PID) {
	pcextr._watchers.Add(watcher)
}

func (pcextr *parallelContextExtras) unwatch(watcher *PID) {
	pcextr._watchers.Remove(watcher)
}

type parallelContext struct {
	_parallel        Parallel
	_system          *ParallelSystem
	_extras          *parallelContextExtras
	_self            *PID
	_receiveTimeout  time.Duration
	_messageOrReport interface{}
	_state           int32
}

func (ctx *parallelContext) instantiateParallel() {
	atomic.StoreInt32(&ctx._state, stateAlive)
	ctx._parallel = ctx._system._opts.Creator(ctx._system)
}

func (ctx *parallelContext) System() *ParallelSystem {
	return ctx._system
}

func (ctx *parallelContext) Logger() logs.LogAgent {
	return ctx._system.Logger()
}

func (ctx *parallelContext) Self() *PID {
	return ctx._self
}

func (ctx *parallelContext) Sender() *PID {
	return UnWrapReportSender(ctx._messageOrReport)
}

func (ctx *parallelContext) Parallel() Parallel {
	return ctx._parallel
}

func (ctx *parallelContext) ReceiveTimeout() time.Duration {
	return ctx._receiveTimeout
}

func (ctx *parallelContext) Respond(response interface{}) {
	// If the message is addressed to nil forward it to the dead letter channel
	if ctx.Sender() == nil {
		ctx._system._deadLetter.postUsrMessage(nil, response)

		return
	}

	ctx.Post(ctx.Sender(), response)
}

func (ctx *parallelContext) Watch(who *PID) {
	who.postSysMessage(ctx._system, &Watch{
		Watcher: ctx._self,
	})
}

func (ctx *parallelContext) Unwatch(who *PID) {
	who.postSysMessage(ctx._system, &Unwatch{
		Watcher: ctx._self,
	})
}

func (ctx *parallelContext) SetReceiveTimeout(d time.Duration) {
	if d <= 0 {
		panic("Duration must be greater than zero")
	}

	if d == ctx._receiveTimeout {
		return
	}

	if d < time.Millisecond {
		// anything less than 1 millisecond is set to zero
		d = 0
	}

	ctx._receiveTimeout = d

	ctx.ensureExtras()
	ctx._extras.stopReceiveTimeoutTimer()

	if d > 0 {
		if ctx._extras._receiveTimeoutTimer == nil {
			ctx._extras.initReceiveTimeoutTimer(time.AfterFunc(d, ctx.receiveTimeoutHandler))
		} else {
			ctx._extras.resetReceiveTimeoutTimer(d)
		}
	}
}

func (ctx *parallelContext) CancelReceiveTimeout() {
	if ctx._extras == nil || ctx._extras._receiveTimeoutTimer == nil {
		return
	}

	ctx._extras.killReceiveTimeoutTimer()
	ctx._receiveTimeout = 0
}

func (ctx *parallelContext) receiveTimeoutHandler() {
	if ctx._extras != nil && ctx._extras._receiveTimeoutTimer != nil {
		ctx.CancelReceiveTimeout()
		ctx.Post(ctx._self, receiveTimeoutMessage)
	}
}

func (ctx *parallelContext) Forward(pid *PID) {
	if msg, ok := ctx._messageOrReport.(SystemMessage); ok {
		// SystemMessage cannot be forwarded
		ctx.Logger().Error("", "SystemMessage cannot be forwarded [message:%+v]", msg)

		return
	}

	ctx.postUsrMessage(pid, ctx._messageOrReport)
}

func (ctx *parallelContext) ReenterAfter(f *Future, cont func(res interface{}, err error)) {
	wrapper := func() {
		cont(f._result, f._err)
	}

	message := ctx._messageOrReport
	// invoke the callback when the future completes
	f.continueWith(func(res interface{}, err error) {
		// send the wrapped callback as a continuation message to self
		ctx._self.postSysMessage(ctx._system, &continuation{
			f:       wrapper,
			message: message,
		})
	})
}

func (ctx *parallelContext) Message() interface{} {
	return UnWrapReportMessage(ctx._messageOrReport)
}

func (ctx *parallelContext) MessageHeader() ReadOnlyMessageHeader {
	return UnWrapReportHeader(ctx._messageOrReport)
}

func (ctx *parallelContext) Post(pid *PID, message interface{}) {
	ctx.postUsrMessage(pid, message)
}

func (ctx *parallelContext) postUsrMessage(pid *PID, message interface{}) {
	pid.postUsrMessage(ctx._system, message)
}

func (ctx *parallelContext) Request(pid *PID, message interface{}) {
	env := &MessageReport{
		Header:  nil,
		Message: message,
		Sender:  ctx.Self(),
	}

	ctx.postUsrMessage(pid, env)
}

func (ctx *parallelContext) RequestWithCustomSender(pid *PID, message interface{}, sender *PID) {
	env := &MessageReport{
		Header:  nil,
		Message: message,
		Sender:  sender,
	}
	ctx.postUsrMessage(pid, env)
}

func (ctx *parallelContext) RequestFuture(pid *PID, message interface{}, timeout time.Duration) *Future {
	future := NewFuture(ctx._system, timeout)
	env := &MessageReport{
		Header:  nil,
		Message: message,
		Sender:  future.PID(),
	}
	ctx.postUsrMessage(pid, env)

	return future
}

func (ctx *parallelContext) Receive(report *MessageReport) {
	ctx._messageOrReport = report
	ctx.defaultReceive()
	ctx._messageOrReport = nil
}

func (ctx *parallelContext) Stop(pid *PID) {
	pid.ref(ctx._system).Stop(pid)
}

func (ctx *parallelContext) StopFuture(pid *PID) *Future {
	future := NewFuture(ctx._system, 10*time.Second)

	pid.postSysMessage(ctx._system, &Watch{Watcher: future._pid})
	ctx.Stop(pid)

	return future
}

// Poison 处理完邮箱中的当前用户消息后会告诉actor停止.
func (ctx *parallelContext) Poison(pid *PID) {
	pid.postSysMessage(ctx._system, poisonPillMessage)
}

func (ctx *parallelContext) PoisonFuture(pid *PID) *Future {
	future := NewFuture(ctx._system, 10*time.Second)

	pid.postSysMessage(ctx._system, &Watch{Watcher: future._pid})
	ctx.Poison(pid)

	return future
}

// Info ...
func (ctx *parallelContext) Info(sfmt string, args ...interface{}) {
	ctx.Logger().Info(fmt.Sprintf("[%s]", ctx._self.ToString()), sfmt, args...)
}

// Debug ...
func (ctx *parallelContext) Debug(sfmt string, args ...interface{}) {
	ctx.Logger().Debug(fmt.Sprintf("[%s]", ctx._self.ToString()), sfmt, args...)
}

// Error ...
func (ctx *parallelContext) Error(sfmt string, args ...interface{}) {
	ctx.Logger().Error(fmt.Sprintf("[%s]", ctx._self.ToString()), sfmt, args...)
}

// Warning ...
func (ctx *parallelContext) Warning(sfmt string, args ...interface{}) {
	ctx.Logger().Warning(fmt.Sprintf("[%s]", ctx._self.ToString()), sfmt, args...)
}

// Fatal ...
func (ctx *parallelContext) Fatal(sfmt string, args ...interface{}) {
	ctx.Logger().Fatal(fmt.Sprintf("[%s]", ctx._self.ToString()), sfmt, args...)
}

// Panic ...
func (ctx *parallelContext) Panic(sfmt string, args ...interface{}) {
	ctx.Logger().Panic(fmt.Sprintf("[%s]", ctx._self.ToString()), sfmt, args...)
}

func (ctx *parallelContext) invokeUsrMessage(md interface{}) {
	if atomic.LoadInt32(&ctx._state) == stateStopped {
		// already stopped
		return
	}

	influenceTimeout := true
	if ctx._receiveTimeout > 0 {
		_, influenceTimeout = md.(NotInfluenceReceiveTimeout)
		influenceTimeout = !influenceTimeout

		if influenceTimeout {
			ctx._extras.stopReceiveTimeoutTimer()
		}
	}

	ctx.handleMessage(md)

	if ctx._receiveTimeout > 0 && influenceTimeout {
		ctx._extras.resetReceiveTimeoutTimer(ctx._receiveTimeout)
	}
}

func (ctx *parallelContext) invokeSysMessage(message interface{}) {
	//goland:noinspection GrazieInspection
	switch msg := message.(type) {
	case *continuation:
		ctx._messageOrReport = msg.message // apply the message that was present when we started the await
		msg.f()                            // invoke the continuation in the current actor context

		ctx._messageOrReport = nil // release the message
	case *Started:
		ctx.invokeUsrMessage(msg) // forward
	case *Watch:
		ctx.onWatch(msg)
	case *Unwatch:
		ctx.onUnwatch(msg)
	case *Stop:
		ctx.onStop(msg)
	case *Terminated:
		ctx.onTerminated(msg)
	case *Failure:
		ctx.onFailure(msg)
	case *Restart:
		ctx.onRestart()
	default:
		ctx.Debug("[PARALLEL] unknown system message [message:%+v]", msg)
	}
}

func (ctx *parallelContext) onWatch(msg *Watch) {
	if atomic.LoadInt32(&ctx._state) >= stateStopping {
		msg.Watcher.postSysMessage(ctx._system, &Terminated{
			Who: ctx._self,
		})
	} else {
		ctx.ensureExtras().watch(msg.Watcher)
	}
}

func (ctx *parallelContext) onUnwatch(msg *Unwatch) {
	if ctx._extras == nil {
		return
	}

	ctx._extras.unwatch(msg.Watcher)
}

func (ctx *parallelContext) onRestart() {
	atomic.StoreInt32(&ctx._state, stateRestarting)
	ctx.invokeUsrMessage(restartingMessage)
	ctx.tryRestartOrTerminate()
}

func (ctx *parallelContext) onStop(message interface{}) {
	if atomic.LoadInt32(&ctx._state) >= stateStopping {
		return
	}

	ctx._state = stateStopping
	atomic.StoreInt32(&ctx._state, stateStopping)
	ctx.invokeUsrMessage(stoppingMessage)
	ctx.tryRestartOrTerminate()
}

func (ctx *parallelContext) onTerminated(message *Terminated) {
	ctx.invokeUsrMessage(message)
	ctx.tryRestartOrTerminate()
}

func (ctx *parallelContext) onFailure(msg *Failure) {
	//if strategy, ok := ctx._parallel.(WatcherStrategy); ok {
	//	strategy.HandleFailure(ctx._system, ctx, msg.Who, msg.RestartStats, msg.Reason, msg.Message)

	//	return
	//}

	//ctx.props.getSupervisor().HandleFailure(ctx.actorSystem, ctx, msg.Who, msg.RestartStats, msg.Reason, msg.Message)
}

func (ctx *parallelContext) tryRestartOrTerminate() {
	if ctx._extras != nil {
		return
	}

	switch atomic.LoadInt32(&ctx._state) {
	case stateRestarting:
		ctx.CancelReceiveTimeout()
		ctx.restart()
	case stateStopping:
		ctx.CancelReceiveTimeout()
		ctx.finalizeStop()
	}
}

func (ctx *parallelContext) restart() {
	ctx.instantiateParallel()
	ctx._self.postSysMessage(ctx._system, resumeMailboxMessage)
	ctx.invokeUsrMessage(startedMessage)
}

func (ctx *parallelContext) finalizeStop() {
	ctx._system._handlers.Erase(ctx._self)
	ctx.invokeSysMessage(stoppedMessage)

	otherStopped := &Terminated{Who: ctx._self}
	// Notify watchers
	if ctx._extras != nil {
		ctx._extras._watchers.ForEach(func(i int, pid *PID) {
			pid.postSysMessage(ctx._system, otherStopped)
		})
	}

	atomic.StoreInt32(&ctx._state, stateStopped)
}

func (ctx *parallelContext) handleMessage(m interface{}) {
	ctx._messageOrReport = m
	ctx.defaultReceive()
	ctx._messageOrReport = nil // release message
}

func (ctx *parallelContext) defaultReceive() {
	switch msg := ctx.Message().(type) {
	case *PoisonPill:
		ctx.Stop(ctx._self)

	case AutoRespond:
		ctx._parallel.Receive(ctx)

		res := msg.GetAutoResponse(ctx)
		ctx.Respond(res)

	default:

		ctx._parallel.Receive(ctx)
	}
}

func (ctx *parallelContext) ensureExtras() *parallelContextExtras {
	if ctx._extras == nil {
		ctx._extras = &parallelContextExtras{}
	}

	return ctx._extras
}

func (ctx *parallelContext) escalateFailure(reason interface{}, message interface{}) {

	ctx.Info("[PARALLEL] Recovering [self:%+v] [reason:%+v]", ctx._self, reason)

	ctx._self.postSysMessage(ctx._system, suspendMailboxMessage)

	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackInfo := fmt.Sprintf("%s", buf[:n])

	ctx.Panic("[PARALLEL] stack:\n %s\n", stackInfo)
}
