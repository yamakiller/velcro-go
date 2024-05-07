package retry

import "fmt"

const maxBackupRetryTimes = 2

// NewBackupPolicy init backup request policy
// the param delayMS is suggested to set as TP99
func NewBackupPolicy(delay uint32) *BackupPolicy {
	if delay == 0 {
		panic("invalid backup request delay duration")
	}
	p := &BackupPolicy{
		RetryDelayMillisecond: delay,
		StopPolicy: StopPolicy{
			MaxRetryTimes:    1,
			DisableChainStop: false,
			CircuitBreakPolicy: CircuitBreakerPolicy{
				ErrorRate: 0.1,
			},
		},
	}
	return p
}

// WithMaxRetryTimes default is 1, not include first call
func (p *BackupPolicy) WithMaxRetryTimes(retryTimes int) {
	if retryTimes < 0 || retryTimes > maxBackupRetryTimes {
		panic(fmt.Errorf("maxBackupRetryTimes=%d", maxBackupRetryTimes))
	}
	p.StopPolicy.MaxRetryTimes = retryTimes
}

// DisableChainRetryStop disables chain stop
func (p *BackupPolicy) DisableChainRetryStop() {
	p.StopPolicy.DisableChainStop = true
}

// WithRetryBreaker sets error rate.
func (p *BackupPolicy) WithRetryBreaker(errRate float64) {
	p.StopPolicy.CircuitBreakPolicy.ErrorRate = errRate
	if err := checkCircuitBreakerErrorRate(&p.StopPolicy.CircuitBreakPolicy); err != nil {
		panic(err)
	}
}

// WithRetrySameNode 设置重试以使用同一节点.
func (p *BackupPolicy) WithRetrySameNode() {
	p.RetrySameNode = true
}

// String 转换为字符串信息
func (p BackupPolicy) String() string {
	return fmt.Sprintf("{RetryDelayMillisecond:%+v StopPolicy:%+v RetrySameNode:%+v}", p.RetryDelayMillisecond, p.StopPolicy, p.RetrySameNode)
}
