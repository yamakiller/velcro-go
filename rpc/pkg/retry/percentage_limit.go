package retry

import "github.com/yamakiller/velcro-go/utils/circuitbreak"

// 将重试视为“错误”以限制重试请求的百分比.
// callTimes == 1 表示这是第一次请求, 而不是重试.
func recordRetryStat(cbrKey string, panel circuitbreak.Panel, callTimes int32) {
	if callTimes > 1 {
		panel.Fail(cbrKey)
		return
	}

	panel.Succeed(cbrKey)
}
