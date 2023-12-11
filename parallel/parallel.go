package parallel

// 生产器
type Creator func() Parallel
type CreatorWithSystem func(system *ParallelSystem) Parallel

type Parallel interface {
	Receive(Context)
}
