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
		"protoclient_client_bytes_receive_duration_seconds",
		metric.WithDescription("Client's bytes received duration in seconds"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientBytesRecviceHistogram instrument, %w", err)
		logger.Error("", err.Error())
	}

	if instruments.ClientBytesSendHistogram, err = meter.Float64Histogram(
		"protoclient_client_bytes_send_duration_seconds",
		metric.WithDescription("Client's bytes send duration in seconds"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientBytesSendHistogram instrument, %w", err)
		logger.Error("", err.Error())
	}

	if instruments.ClientMessageRecviceHistogram, err = meter.Int64Histogram(
		"protoclient_client_message_recvice_duration_seconds",
		metric.WithDescription("Client's message send duration in seconds"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientMessageRecviceHistogram instrument, %w", err)
		logger.Error("", err.Error())
	}

	if instruments.ClientMessageSendHistogram, err = meter.Int64Histogram(
		"protoclient_client_message_send_duration_seconds",
		metric.WithDescription("Client's message send duration in seconds"),
	); err != nil {
		err = fmt.Errorf("failed to create ClientMessageSendHistogram instrument, %w", err)
		logger.Error("", err.Error())
	}

	return &instruments
}
