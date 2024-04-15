package apps

// import (
// 	"fmt"
// 	"net/http"

// 	"github.com/prometheus/client_golang/prometheus/promhttp"
// 	"go.opentelemetry.io/otel"
// 	"go.opentelemetry.io/otel/exporters/prometheus"
// 	"go.opentelemetry.io/otel/metric"

// 	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
// )

// // 用于本地调试用，后面可以注释掉
// func localPrometheusProvider(port int) metric.MeterProvider {
// 	exporter, err := prometheus.New()
// 	if err != nil {
// 		err = fmt.Errorf("failed to initialize prometheus exporter: %w", err)
// 		return nil
// 	}

// 	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter.Reader))
// 	otel.SetMeterProvider(provider)

// 	http.Handle("/", promhttp.Handler())
// 	_port := fmt.Sprintf(":%d", port)

// 	go func() {
// 		_ = http.ListenAndServe(_port, nil)
// 	}()

// 	return provider
// }
