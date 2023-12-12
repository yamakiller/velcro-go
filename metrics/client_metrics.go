package metrics

import (
	"fmt"
	"sync"

	"github.com/yamakiller/velcro-go/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const LibName string = "protoclient"

type ClientMetrics struct {
	_mutex *sync.Mutex

	// MetricsID
	ID string

	//
	ClientBytesRecviceHistogram   metric.Float64Histogram
	ClientBytesSendHistogram      metric.Float64Histogram
	ClientMessageRecviceHistogram metric.Int64Histogram
	ClientMessageSendHistogram    metric.Int64Histogram

	ClientSpawnCount metric.Int64Counter
	ClientCloseCount metric.Int64Counter

	// Threadpool
	//ThreadPoolLatency metric.Int64Histogram
}

func NewClientMetrics(logger logs.LogAgent) *ClientMetrics {
	instruments := newInstruments(logger)
	return instruments
}

func newInstruments(logger logs.LogAgent) *ClientMetrics {

	meter := otel.Meter(LibName)
	instruments := ClientMetrics{_mutex: &sync.Mutex{}}

	var err error

	if instruments.ClientBytesRecviceHistogram, err = meter.Float64Histogram(
		"protonetwork_client_bytes_receive_duration_seconds",
		metric.WithDescription("Client's bytes received duration in seconds"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientBytesRecviceHistogram instrument, %w", err)
		logger.Error("[METRICS]", err.Error())
	}

	if instruments.ClientBytesSendHistogram, err = meter.Float64Histogram(
		"protonetwork_client_bytes_send_duration_seconds",
		metric.WithDescription("Client's bytes send duration in seconds"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientBytesSendHistogram instrument, %w", err)
		logger.Error("[METRICS]", err.Error())
	}

	if instruments.ClientMessageRecviceHistogram, err = meter.Int64Histogram(
		"protonetwork_client_message_recvice_duration_seconds",
		metric.WithDescription("Client's message send duration in seconds"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientMessageRecviceHistogram instrument, %w", err)
		logger.Error("[METRICS]", err.Error())
	}

	if instruments.ClientMessageSendHistogram, err = meter.Int64Histogram(
		"protonetwork_client_message_send_duration_seconds",
		metric.WithDescription("Client's message send duration in seconds"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientMessageSendHistogram instrument, %w", err)
		logger.Error("[METRICS]", err.Error())
	}

	if instruments.ClientSpawnCount, err = meter.Int64Counter(
		"protonetwork_client_spawn_count",
		metric.WithDescription("Number of client spawn"),
		metric.WithUnit("1"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientSpawnCount instrument, %w", err)
		logger.Error("[METRICS]", err.Error())
	}

	if instruments.ClientCloseCount, err = meter.Int64Counter(
		"protonetwork_client_close_count",
		metric.WithDescription("Number of client closed"),
		metric.WithUnit("1"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientCloseCount instrument, %w", err)
		logger.Error("[METRICS]", err.Error())
	}

	return &instruments
}
