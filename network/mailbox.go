package network

/*const (
	idle int32 = iota
	running
)

type MessageInvoker interface {
	invokeSysMessage(interface{})
	invokeUsrMessage(interface{})
	EscalateFailure(reason interface{}, message interface{})
}

type mailbox struct {
	_usrMailbox      *containers.Queue
	_sysMailbox      *containers.Queue
	_usrMessages     int32
	_sysMessages     int32
	_invoker         MessageInvoker
	_dispatcher      dispatcher
	_schedulerStatus int32
}

func (m *mailbox) posUsrMessage(message interface{}) {
	m._usrMailbox.Push(message)
	atomic.AddInt32(&m._usrMessages, 1)
	m.schedule()
}

func (m *mailbox) schedule() {
	if atomic.CompareAndSwapInt32(&m._schedulerStatus, idle, running) {
		m._dispatcher.Schedule(m.processMessages)
	}
}

func (m *mailbox) processMessages() {
process_lable:
	m.run()

	atomic.StoreInt32(&m._schedulerStatus, idle)
	sys := atomic.LoadInt32(&m._sysMessages)
	usr := atomic.LoadInt32(&m._usrMessages)

	if sys > 0 || usr > 0 {
		if atomic.CompareAndSwapInt32(&m._schedulerStatus, idle, running) {
			goto process_lable
		}
	}
}

func (m *mailbox) run() {
	var msg interface{}
	var ok bool

	defer func() {
		if r := recover(); r != nil {
			m._invoker.EscalateFailure(r, msg)
		}
	}()

	i, t := 0, m._dispatcher.Throughput()
	for {
		if i > t {
			i = 0
			runtime.Gosched()
		}

		i++

		if msg, ok = m._sysMailbox.Pop(); ok {
			atomic.AddInt32(&m._sysMessages, -1)
			m._invoker.invokeSysMessage(msg)
			continue
		}

		if msg, ok = m._usrMailbox.Pop(); ok {
			atomic.AddInt32(&m._usrMessages, -1)
			m._invoker.invokeUsrMessage(msg)
		}
	}
}
*/
