package gopool

const (
	defaultScalaThreshold = 1
)

// Config is used to config pool.
type Config struct {
	// 协程池阈值
	ScaleThreshold int32
}

// NewConfig creates a default Config.
func NewConfig() *Config {
	c := &Config{
		ScaleThreshold: defaultScalaThreshold,
	}
	return c
}
