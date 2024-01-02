package limiter

var defaultLimiterConfig = &LimiterConfig{}

// LimiterConfig represents the configuration for kitex server limiter
// zero value means no limit.
type LimiterConfig struct {
	ConnectionLimit int64 `yaml:"connection_limit"`
	QPSLimit        int64 `yaml:"qps_limit"`
}
