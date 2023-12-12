package network

import (
	"fmt"
	"strings"

	"github.com/yamakiller/velcro-go/extensions"
	"github.com/yamakiller/velcro-go/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var extensionId = extensions.NextExtensionID()

type Metrics struct {
	_metrics *metrics.ProtoMetrics
	_enabled bool
	_system  *NetworkSystem
}

var _ extensions.Extension = &Metrics{}

func (m *Metrics) Enabled() bool {
	return m._enabled
}

func (m *Metrics) ExtensionID() extensions.ExtensionID {
	return extensionId
}

func NewMetrics(system *NetworkSystem, provider metric.MeterProvider) *Metrics {
	if provider == nil {
		return &Metrics{}
	}

	return &Metrics{
		_metrics: metrics.NewProtoMetrics(system.Logger()),
		_enabled: true,
		_system:  system,
	}
}

func (m *Metrics) CommonLabels(ctx Context) []attribute.KeyValue {
	labels := []attribute.KeyValue{
		attribute.String("address", ctx.NetworkSystem().Address()),
		attribute.String("clienttype", strings.Replace(fmt.Sprintf("%T", ctx.Client()), "*", "", 1)),
	}

	return labels
}
