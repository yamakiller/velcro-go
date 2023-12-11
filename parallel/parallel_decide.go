package parallel

// Decide 决定让守护者对守护的并行器失败后作出什么样的处理
type Decide int

const (
	// ResumeDirective 让并行器自动回复运行
	ResumeDirective Decide = iota

	// RestartDirective 让并行器重新启动
	RestartDirective

	// StopDirective 停止并行器
	StopDirective

	// EscalateDirective 报告故障信息
	EscalateDirective
)
