package limiter

type LimiterConfig struct {
	ConnectionLimiter int64 `json:"connection limit"`
	QPSLimit          int64 `json:"qps_limit"`
}

func (l *LimiterConfig) EqualsTo(other interface{}) bool {
	o := other.(*LimiterConfig)
	return l.ConnectionLimiter == o.ConnectionLimiter && l.QPSLimit == o.QPSLimit
}
