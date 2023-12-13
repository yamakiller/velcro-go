package metrics

import (
	"fmt"
	"sync"

	"github.com/yamakiller/velcro-go/logs"
)

const InternalClientMetrics string = "internal.client.metrics"

type ProtoMetrics struct {
	_mutex         sync.Mutex
	_clientMetrics *ClientMetrics
	_knownMetrics  map[string]*ClientMetrics
	_logger        logs.LogAgent
}

func NewProtoMetrics(logger logs.LogAgent) *ProtoMetrics {
	protoMetrics := ProtoMetrics{
		_clientMetrics: NewClientMetrics(logger),
		_knownMetrics:  make(map[string]*ClientMetrics),
		_logger:        logger,
	}

	protoMetrics.Register(InternalClientMetrics, protoMetrics._clientMetrics)
	return &protoMetrics
}

func (pm *ProtoMetrics) Instruments() *ClientMetrics { return pm._clientMetrics }

func (pm *ProtoMetrics) Register(key string, instance *ClientMetrics) {
	pm._mutex.Lock()
	defer pm._mutex.Unlock()
	logger := pm._logger

	if _, ok := pm._knownMetrics[key]; ok {
		err := fmt.Errorf("could not register instance %#v of metrics, %s already registered", instance, key)
		logger.Error("", err.Error())
		return
	}

	pm._knownMetrics[key] = instance
}

func (pm *ProtoMetrics) Get(key string) *ClientMetrics {
	metrics, ok := pm._knownMetrics[key]
	if !ok {
		logger := pm._logger
		err := fmt.Errorf("unknown metrics for the given %s key", key)
		logger.Error("", err.Error())
		return nil
	}

	return metrics
}
