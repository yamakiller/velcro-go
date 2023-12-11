package parallel

import "sync/atomic"

type Handler interface {
	overloadUsrMessage() int
	postUsrMessage(pid *PID, message interface{})
	postSysMessage(pid *PID, message interface{})
	Stop(pid *PID)
}

type ParallelHandler struct {
	_mailbox mailbox
	_death   int32
}

// overloadUsrMessage Returns user message overload warring
func (slf *ParallelHandler) overloadUsrMessage() int {
	return slf._mailbox._usrMailbox.Overload()
}

// postUsrMessage Send user level messages
func (slf *ParallelHandler) postUsrMessage(_ *PID, message interface{}) {
	slf._mailbox.postUsrMessage(message)
}

// postSysMessage Send system level messages
func (slf *ParallelHandler) postSysMessage(_ *PID, message interface{}) {
	slf._mailbox.postSysMessage(message)
}

// Stop Send stop Actor message
func (slf *ParallelHandler) Stop(pid *PID) {
	atomic.StoreInt32(&slf._death, 1)
	slf.postSysMessage(pid, stopMessage)
}
