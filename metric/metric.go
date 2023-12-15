package metric

type Opts struct {
}

// Metric 一个样本接口
type Metric interface {
	// Add 将给定值添加到计数器.
	Add(int64)
	// Value 获取当前值.
	// 如果度量类型为 PointGauge、RollingCounter、RollingGauge, 则返回窗口内的总和值.
	Value() int64
}

// Aggregation 包含一些常见的聚合函数. 每个聚合可以计算窗口的汇总统计数据.
type Aggregation interface {
	// Min finds the min value within the window.
	Min() float64
	// Max finds the max value within the window.
	Max() float64
	// Avg computes average value within the window.
	Avg() float64
	// Sum computes sum value within the window.
	Sum() float64
}

// VectorOpts 包含创建 vec Metric 的常见参数.
type VectorOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

const (
	_businessNamespace          = "business"
	_businessSubsystemCount     = "count"
	_businessSubSystemGauge     = "gauge"
	_businessSubSystemHistogram = "histogram"
)

var (
	_defaultBuckets = []float64{5, 10, 25, 50, 100, 250, 500}
)
