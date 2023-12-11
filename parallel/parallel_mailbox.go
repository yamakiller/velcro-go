package parallel

import (
	"runtime"
	"sync/atomic"

	"github.com/yamakiller/velcro-go/containers"
)

const (
	idle int32 = iota
	running
)

const (
	PriorityNomal int32 = iota
	PriorityHigh
)

type mailbox struct {
	_usrMailbox      *containers.Queue
	_sysMailbox      *containers.Queue
	_usrMessages     int32
	_sysMessages     int32
	_schedulerStatus int32
	_priority        int32
	_dispatcher      dispatcher
	_invoker         invoker
	_suspended       int32
}

func (m *mailbox) RegisterHandlers(invoker invoker, dispatcher dispatcher) {
	m._invoker = invoker
	m._dispatcher = dispatcher
}

func (m *mailbox) postSysMessage(msg interface{}) {
	m._sysMailbox.Push(msg)
	atomic.AddInt32(&m._sysMessages, 1)
	m.schedule()
}

func (m *mailbox) postUsrMessage(msg interface{}) {
	if batch, ok := msg.(MessageBatch); ok {
		messages := batch.GetMessages()

		for _, msg := range messages {
			m.postUsrMessage(msg)
		}
	}

	if env, ok := msg.(MessageReport); ok {
		if batch, ok := env.Message.(MessageBatch); ok {
			messages := batch.GetMessages()

			for _, msg := range messages {
				m.postUsrMessage(msg)
			}
		}
	}

	if env, ok := msg.(*MessageReport); ok {
		if batch, ok := env.Message.(MessageBatch); ok {
			messages := batch.GetMessages()

			for _, msg := range messages {
				m.postUsrMessage(msg)
			}
		}
	}

	m._usrMailbox.Push(msg)
	atomic.AddInt32(&m._usrMessages, 1)
	m.schedule()
}

func (m *mailbox) schedule() {
	if atomic.CompareAndSwapInt32(&m._schedulerStatus, idle, running) {
		m._dispatcher.Schedule(m.processMessages)
	}
}

func (m *mailbox) processMessages([]interface{}) {
process_lable:
	m.run()

	atomic.StoreInt32(&m._schedulerStatus, idle)
	sys := atomic.LoadInt32(&m._sysMessages)
	usr := atomic.LoadInt32(&m._usrMessages)

	if sys > 0 || (atomic.LoadInt32(&m._suspended) == 0 && usr > 0) {
		if atomic.CompareAndSwapInt32(&m._schedulerStatus, idle, running) {
			goto process_lable
		}
	}
}

func (m *mailbox) run() {
	var msg interface{}
	//Begin=>致命性异常处理
	defer func() {
		if r := recover(); r != nil {
			m._invoker.escalateFailure(r, msg)
		}
	}()
	//End=>致命性异常处理结束
	i := 0
	for {
		if i > 0 && m._priority == PriorityNomal {
			i = 0
			runtime.Gosched()
		}

		i++
		if msg, _ = m._sysMailbox.Pop(); msg != nil {
			atomic.AddInt32(&m._sysMessages, -1)
			switch msg.(type) {
			case *SuspendMailbox:
				atomic.StoreInt32(&m._suspended, 1)
			case ResumeMailbox:
				atomic.StoreInt32(&m._suspended, 0)
			default:
				m._invoker.invokeSysMessage(msg)
			}
			continue
		}

		if atomic.LoadInt32(&m._suspended) == 1 {
			return
		}

		if msg, _ = m._usrMailbox.Pop(); msg != nil {
			atomic.AddInt32(&m._usrMessages, -1)
			m._invoker.invokeUsrMessage(msg)
		} else {
			return
		}

	}
}
