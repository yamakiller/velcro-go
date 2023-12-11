package parallel

type MailboxMessage interface {
	MailboxMessage()
}

func (*SuspendMailbox) MailboxMessage() {}
func (*ResumeMailbox) MailboxMessage()  {}

// InfrastructureMessage 是所有内置 Proto.Actor 消息的标记
type InfrastructureMessage interface {
	InfrastructureMessage()
}

// IgnoreDeadLetterLogging 消息未记录在死信日志中
type IgnoreDeadLetterLogging interface {
	IgnoreDeadLetterLogging()
}

// An AutoReceiveMessage 是一种特殊的用户消息，将由参与者以某种方式自动处理
type AutoReceiveMessage interface {
	AutoReceiveMessage()
}

// NotInfluenceReceiveTimeout 消息不会重置接收消息的 Actor 的 ReceiveTimeout 计时器
type NotInfluenceReceiveTimeout interface {
	NotInfluenceReceiveTimeout()
}

// A SystemMessage message 为参与者系统使用的特定生命周期消息保留
type SystemMessage interface {
	SystemMessage()
}

type Failure struct {
	Who          *PID
	Reason       interface{}
	RestartStats *RestartStatistics
	Message      interface{}
}

type continuation struct {
	message interface{}
	f       func()
}

func (*Touch) GetAutoResponse(ctx Context) interface{} {
	return &Touched{
		Who: ctx.Self(),
	}
}

func (*Restarting) AutoReceiveMessage() {}
func (*Stopping) AutoReceiveMessage()   {}
func (*Stopped) AutoReceiveMessage()    {}
func (*PoisonPill) AutoReceiveMessage() {}

func (*Started) SystemMessage()      {}
func (*Stop) SystemMessage()         {}
func (*Watch) SystemMessage()        {}
func (*Unwatch) SystemMessage()      {}
func (*Terminated) SystemMessage()   {}
func (*Failure) SystemMessage()      {}
func (*Restart) SystemMessage()      {}
func (*continuation) SystemMessage() {}

var (
	restartingMessage     AutoReceiveMessage = &Restarting{}
	stoppingMessage       AutoReceiveMessage = &Stopping{}
	stoppedMessage        AutoReceiveMessage = &Stopped{}
	poisonPillMessage     AutoReceiveMessage = &PoisonPill{}
	receiveTimeoutMessage interface{}        = &ReceiveTimeout{}
	restartMessage        SystemMessage      = &Restart{}
	startedMessage        SystemMessage      = &Started{}
	stopMessage           SystemMessage      = &Stop{}
	resumeMailboxMessage  MailboxMessage     = &ResumeMailbox{}
	suspendMailboxMessage MailboxMessage     = &SuspendMailbox{}
	_                     AutoRespond        = &Touch{}
)
