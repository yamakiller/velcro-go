package parallel

import "time"

// RestartStatisitics 重启统计数据
// 跟踪并行器重新启动的次数和时间
type RestartStatistics struct {
	_failureTimes []time.Time // 失败时间
}

// NewRestartStatistics 分配一个重启统计对象
func NewRestartStatistics() *RestartStatistics {
	return &RestartStatistics{[]time.Time{}}
}

// FailureCount 失败次数
func (rs *RestartStatistics) FailureCount() int {
	return len(rs._failureTimes)
}

// Fail 增加一次失败
func (rs *RestartStatistics) Fail() {
	rs._failureTimes = append(rs._failureTimes, time.Now())
}

// Reset 重置统计数据
func (rs *RestartStatistics) Reset() {
	rs._failureTimes = []time.Time{}
}

// NumberOfFailures 统计在时间段内失败的次数
func (rs *RestartStatistics) NumberOfFailures(withinDuration time.Duration) int {
	if withinDuration == 0 {
		return len(rs._failureTimes)
	}

	num := 0
	currTime := time.Now()

	for _, t := range rs._failureTimes {
		if currTime.Sub(t) < withinDuration {
			num++
		}
	}

	return num
}
