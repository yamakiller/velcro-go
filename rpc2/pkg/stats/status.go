package stats

type Status int8

// 预定义状态
const (
	StatusInfo  Status = 1
	StatusWarn  Status = 2
	StatusError Status = 3
)
