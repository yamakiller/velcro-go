package network

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/yamakiller/velcro-go/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Config struct {
	MetricsProvider metric.MeterProvider
	MeriicsKey      string
	LoggerFactory   func(system *NetworkSystem) logs.LogAgent // 日志仓库
	Producer        ProducerWidthClientSystem
	NetowkTimeout   int32 //网络超时(单位毫秒)
}

func defaultConfig() *Config {
	return &Config{
		MetricsProvider: nil, LoggerFactory: func(system *NetworkSystem) logs.LogAgent {
			pLogHandle := logs.SpawnFileLogrus(logrus.TraceLevel, "", "Proto.Network."+system.NetType+"."+system.ID+"."+system.Address())

			logAgent := &logs.DefaultAgent{}
			logAgent.WithHandle(pLogHandle)
			return logAgent
		},
		NetowkTimeout: 2000}
}

func defaultPrometheusProvider(port int) metric.MeterProvider {
	exporter, _ := prometheus.New()
	/*if err != nil {
		err = fmt.Errorf("failed to initialize prometheus exporter: %w", err)
		return nil
	}*/

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter.Reader))
	otel.SetMeterProvider(provider)

	http.Handle("/", promhttp.Handler())
	_port := fmt.Sprintf(":%d", port)

	go func() {
		_ = http.ListenAndServe(_port, nil)
	}()

	return provider
}
