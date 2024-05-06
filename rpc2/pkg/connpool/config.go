package connpool

import "time"

// IdleConfig 包含长连接池的空闲配置.
type IdleConfig struct {
	MinIdlePerAddress int
	MaxIdlePerAddress int
	MaxIdleGlobal     int
	MaxIdleTimeout    time.Duration
}

const (
	defaultMaxIdleTimeout       = 30 * time.Second
	defaultMinMaxIdleTimeout    = 2 * time.Second
	defaultMaxMinIdlePerAddress = 5
	defaultMaxIdleGlobal        = 1 << 20
)

// VerifyPoolConfig 用于校验并修正配置错误
func VerifyPoolConfig(config IdleConfig) *IdleConfig {
	// idle timeout
	if config.MaxIdleTimeout == 0 {
		config.MaxIdleTimeout = defaultMaxIdleTimeout
	} else if config.MaxIdleTimeout < defaultMinMaxIdleTimeout {
		config.MaxIdleTimeout = defaultMinMaxIdleTimeout
	}

	// idlePerAddress
	if config.MinIdlePerAddress < 0 {
		config.MinIdlePerAddress = 0
	}
	if config.MinIdlePerAddress > defaultMaxMinIdlePerAddress {
		config.MinIdlePerAddress = defaultMaxMinIdlePerAddress
	}
	if config.MaxIdlePerAddress <= 0 {
		config.MaxIdlePerAddress = 1
	}
	if config.MaxIdlePerAddress < config.MinIdlePerAddress {
		config.MaxIdlePerAddress = config.MinIdlePerAddress
	}

	// globalIdle
	if config.MaxIdleGlobal <= 0 {
		config.MaxIdleGlobal = defaultMaxIdleGlobal
	} else if config.MaxIdleGlobal < config.MaxIdlePerAddress {
		config.MaxIdleGlobal = config.MaxIdlePerAddress
	}
	return &config
}
