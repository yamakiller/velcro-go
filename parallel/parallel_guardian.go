package parallel

// 并行器守护

type guardian struct {
	_system   ParallelSystem
	_strategy GuardStrategy
}

type GuardStrategy interface {
}
