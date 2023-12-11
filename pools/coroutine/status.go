package coroutine

// State 运行状态 enum
type State int32

const (
	// Idle 协程处于闲置状态
	IdleState State = iota
	// Run  协程处于运行状态
	RunState
	// Dead 协程已死亡
	DeadState
)
