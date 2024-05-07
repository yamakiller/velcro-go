package retry

import (
	"fmt"

	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
)

// Type is retry type include FailureType, BackupType
type Type int

// retry types
const (
	FailureType Type = iota
	BackupType
)

// String 输出字符串信息
func (t Type) String() string {
	switch t {
	case FailureType:
		return "Failure"
	case BackupType:
		return "Backup"
	}
	return ""
}

// SpwnFailurePolicy is used to build Policy with *FailurePolicy
func SpwnFailurePolicy(p *FailurePolicy) Policy {
	if p == nil {
		return Policy{}
	}
	return Policy{Enable: true, Type: FailureType, FailurePolicy: p}
}

// SpwnBackupRequest is used to build Policy with *BackupPolicy
func SpwnBackupRequest(p *BackupPolicy) Policy {
	if p == nil {
		return Policy{}
	}
	return Policy{Enable: true, Type: BackupType, BackupPolicy: p}
}

// Policy 包含所有重试策略
// 如果添加新字段, 请不要忘记更新 Equals() 和 DeepCopy()
type Policy struct {
	Enable bool `yaml:"enable"`
	// 0 is failure retry, 1 is backup
	Type Type `yaml:"type"`
	// 注意:只能启用一项重试策略,具体取决于Policy.Type
	FailurePolicy *FailurePolicy `yaml:"failure_policy,omitempty"`
	BackupPolicy  *BackupPolicy  `yaml:"backup_policy,omitempty"`
}

func (p *Policy) DeepCopy() *Policy {
	if p == nil {
		return nil
	}
	return &Policy{
		Enable:        p.Enable,
		Type:          p.Type,
		FailurePolicy: p.FailurePolicy.DeepCopy(),
		BackupPolicy:  p.BackupPolicy.DeepCopy(),
	}
}

// Equals to check if policy is equal
func (p Policy) Equals(np Policy) bool {
	if p.Enable != np.Enable {
		return false
	}
	if p.Type != np.Type {
		return false
	}
	if !p.FailurePolicy.Equals(np.FailurePolicy) {
		return false
	}
	if !p.BackupPolicy.Equals(np.BackupPolicy) {
		return false
	}
	return true
}

type FailurePolicy struct {
	StopPolicy        StopPolicy         `yaml:"stop_policy"`
	BackOffPolicy     *BackOffPolicy     `yaml:"backoff_policy,omitempty"`
	RetrySameNode     bool               `yaml:"retry_same_node"`
	ShouldResultRetry *ShouldResultRetry `yaml:"-"`

	// Addon velcro 不直接使用. 它用于更好地集成您自己的配置源.
	// 从配置源加载 FailurePolicy 后, 'Addon' 可以解码为用户定义的模式,
	// 通过它，可以实施更复杂的策略，例如修改 'ShouldResultRetry'.
	Addon string `yaml:"addon"`
}

// IsRespRetryNonNil is used to check if RespRetry is nil
func (p FailurePolicy) IsRespRetryNonNil() bool {
	return p.ShouldResultRetry != nil && p.ShouldResultRetry.RespRetry != nil
}

// IsErrorRetryNonNil is used to check if ErrorRetry is nil
func (p FailurePolicy) IsErrorRetryNonNil() bool {
	return p.ShouldResultRetry != nil && p.ShouldResultRetry.ErrorRetry != nil
}

// IsRetryForTimeout is used to check if timeout error need to retry
func (p FailurePolicy) IsRetryForTimeout() bool {
	return p.ShouldResultRetry == nil || !p.ShouldResultRetry.NotRetryForTimeout
}

// Equals to check if FailurePolicy is equal
func (p *FailurePolicy) Equals(np *FailurePolicy) bool {
	if p == nil {
		return np == nil
	}
	if np == nil {
		return false
	}
	if p.StopPolicy != np.StopPolicy {
		return false
	}
	if !p.BackOffPolicy.Equals(np.BackOffPolicy) {
		return false
	}
	if p.RetrySameNode != np.RetrySameNode {
		return false
	}
	if p.Addon != np.Addon {
		return false
	}
	// don't need to check `ShouldResultRetry`, ShouldResultRetry is only setup by option
	// in remote config case will always return false if check it
	return true
}

func (p *FailurePolicy) DeepCopy() *FailurePolicy {
	if p == nil {
		return nil
	}
	return &FailurePolicy{
		StopPolicy:        p.StopPolicy,
		BackOffPolicy:     p.BackOffPolicy.DeepCopy(),
		RetrySameNode:     p.RetrySameNode,
		ShouldResultRetry: p.ShouldResultRetry, // don't need DeepCopy
		Addon:             p.Addon,
	}
}

// BackupPolicy 用于回退请求
// 如果添加新字段, 请不要忘记更新 Equals() 和 DeepCopy().
type BackupPolicy struct {
	RetryDelayMillisecond uint32     `yaml:"retry_delay_millisecond"`
	StopPolicy            StopPolicy `yaml:"stop_policy"`
	RetrySameNode         bool       `yaml:"retry_same_node"`
}

// Equals to check if BackupPolicy is equal
func (p *BackupPolicy) Equals(np *BackupPolicy) bool {
	if p == nil {
		return np == nil
	}
	if np == nil {
		return false
	}
	if p.RetryDelayMillisecond != np.RetryDelayMillisecond {
		return false
	}
	if p.StopPolicy != np.StopPolicy {
		return false
	}
	if p.RetrySameNode != np.RetrySameNode {
		return false
	}

	return true
}

func (p *BackupPolicy) DeepCopy() *BackupPolicy {
	if p == nil {
		return nil
	}
	return &BackupPolicy{
		RetryDelayMillisecond: p.RetryDelayMillisecond,
		StopPolicy:            p.StopPolicy, // not a pointer, will copy the value here
		RetrySameNode:         p.RetrySameNode,
	}
}

// StopPolicy 是决定何时停止重试的组策略
type StopPolicy struct {
	MaxRetryTimes          int                  `yaml:"max_retry_times"`
	MaxDurationMillisecond uint32               `yaml:"max_duration_millisecond"`
	DisableChainStop       bool                 `yaml:"disable_chain_stop"`
	DDLStop                bool                 `yaml:"ddl_stop"`
	CircuitBreakPolicy     CircuitBreakerPolicy `yaml:"circuitbreaker_policy"`
}

const (
	defaultCircuitBreakerErrRate = 0.1
	circuitBreakerMinSample      = 10
)

// CircuitBreakerPolicy 断路策略
type CircuitBreakerPolicy struct {
	ErrorRate float64 `yaml:"error_rate"`
}

func checkCircuitBreakerErrorRate(p *CircuitBreakerPolicy) error {
	if p.ErrorRate <= 0 || p.ErrorRate > 0.3 {
		return fmt.Errorf("invalid retry circuit breaker rate, errRate=%0.2f", p.ErrorRate)
	}
	return nil
}

type BackOffPolicy struct {
	BackOffType BackOffType               `yaml:"backoff_type"`
	CfgItems    map[BackOffCfgKey]float64 `yaml:"cfg_items, omitempty"`
}

// Equals to check if BackOffPolicy is equal.
func (p *BackOffPolicy) Equals(np *BackOffPolicy) bool {
	if p == nil {
		return np == nil
	}
	if np == nil {
		return false
	}
	if p.BackOffType != np.BackOffType {
		return false
	}
	if len(p.CfgItems) != len(np.CfgItems) {
		return false
	}
	for k := range p.CfgItems {
		if p.CfgItems[k] != np.CfgItems[k] {
			return false
		}
	}

	return true
}

func (p *BackOffPolicy) DeepCopy() *BackOffPolicy {
	if p == nil {
		return nil
	}
	return &BackOffPolicy{
		BackOffType: p.BackOffType,
		CfgItems:    p.copyCfgItems(),
	}
}

func (p *BackOffPolicy) copyCfgItems() map[BackOffCfgKey]float64 {
	if p.CfgItems == nil {
		return nil
	}
	cfgItems := make(map[BackOffCfgKey]float64, len(p.CfgItems))
	for k, v := range p.CfgItems {
		cfgItems[k] = v
	}
	return cfgItems
}

// BackOffType 表示 BackOff 类型.
type BackOffType string

// all back off types
const (
	NoneBackOffType   BackOffType = "none"
	FixedBackOffType  BackOffType = "fixed"
	RandomBackOffType BackOffType = "random"
)

// BackOffCfgKey represents the keys for BackOff.
type BackOffCfgKey string

// 回退开关配置关键字
const (
	FixMSBackOffCfgKey      BackOffCfgKey = "fix_ms"
	MinMSBackOffCfgKey      BackOffCfgKey = "min_ms"
	MaxMSBackOffCfgKey      BackOffCfgKey = "max_ms"
	InitialMSBackOffCfgKey  BackOffCfgKey = "initial_ms"
	MultiplierBackOffCfgKey BackOffCfgKey = "multiplier"
)

// ShouldResultRetry 用于指定需要重试哪个错误或响应
type ShouldResultRetry struct {
	ErrorRetry func(err error, ri rpcinfo.RPCInfo) bool
	RespRetry  func(resp interface{}, ri rpcinfo.RPCInfo) bool
	// 特定场景禁用默认超时重试 (e.g. the requests are not non-idempotent)
	NotRetryForTimeout bool
}
