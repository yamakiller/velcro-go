package limit

// Updater is used to update the limit dynamically.
type Updater interface {
	UpdateLimit(opt *Option) (updated bool)
}

// Option 限制设置
type Option struct {
	MaxConnections int // 最大连接数
	MaxQPS         int // 最大QPS

	// UpdateControl receives a Updater which gives the limitation provider
	// the ability to update limit dynamically.
	UpdateControl func(u Updater)
}

// Valid 检测限制设置是否有效
func (lo *Option) Valid() bool {
	return lo.MaxConnections > 0 || lo.MaxQPS > 0
}
