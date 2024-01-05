package clientpool

import "time"

type IdleConfig struct {
	MaxIdleGlobal      int32
	MaxIdleTimeout     time.Duration
	MaxIdleConnTimeout time.Duration

	Kleepalive int32

	Connected func()
	Closed    func()
}

const (
	defaultMaxIdleTimeout = 30 * time.Second
	minMaxIdleTimeout     = 2 * time.Second
	maxIdleConnTimeout    = 2 * time.Second
	defaultMaxIdleGlobal  = 1 << 20 // no limit
)

// CheckPoolConfig to check invalid param.
// default MaxIdleTimeout = 30s, min value is 2s
func CheckPoolConfig(config IdleConfig) *IdleConfig {
	// idle timeout
	if config.MaxIdleTimeout == 0 {
		config.MaxIdleTimeout = defaultMaxIdleTimeout
	} else if config.MaxIdleTimeout < minMaxIdleTimeout {
		config.MaxIdleTimeout = minMaxIdleTimeout
	}

	// idlePerAddress
	if config.MaxIdleConnTimeout == 0 {
		config.MaxIdleConnTimeout = maxIdleConnTimeout
	}

	// globalIdle
	if config.MaxIdleGlobal <= 0 {
		config.MaxIdleGlobal = defaultMaxIdleGlobal
	}

	return &config
}
