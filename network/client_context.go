package network

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// 通用客户端上下文
type clientContext struct {
	client      Client
	system      *NetworkSystem
	self        *ClientID
	message     []byte
	messageFrom net.Addr
	state       int32
}

func (ctx *clientContext) Client() Client {
	return ctx.client
}

func (ctx *clientContext) Self() *ClientID {
	return ctx.self
}

func (ctx *clientContext) NetworkSystem() *NetworkSystem {
	return ctx.system
}

// Message 返回当前接收到的消息
func (ctx *clientContext) Message() []byte {
	return ctx.message
}

func (ctx *clientContext) MessageFrom() net.Addr {
	return ctx.messageFrom
}

func (ctx *clientContext) PostMessage(cid *ClientID, message []byte) error {
	return cid.postMessage(ctx.system, message)
}

func (ctx *clientContext) PostToMessage(cid *ClientID, message []byte, target net.Addr) error {
	return cid.postToMessage(ctx.system, message, target)
}

// Close 关闭当前 Client
func (ctx *clientContext) Close(cid *ClientID) {
	if ctx.system.Config.MetricsProvider != nil {
		metricsSystem, ok := ctx.system.extensions.Get(ctx.system.extensionId).(*Metrics)
		if ok && metricsSystem._enabled {
			_ctx := context.Background()
			if instruments := metricsSystem._metrics.Get(ctx.system.Config.meriicsKey); instruments != nil {
				instruments.ClientCloseCount.Add(_ctx, 1, metric.WithAttributes(metricsSystem.CommonLabels(ctx)...))
			}
		}
	}

	cid.close(ctx.system)
}

func (ctx *clientContext) incarnateClient() {
	atomic.StoreInt32(&ctx.state, stateAccept)

	ctx.client = ctx.system.producer(ctx.system)

	if ctx.system.Config.MetricsProvider != nil {
		metricsSystem, ok := ctx.system.extensions.Get(ctx.system.extensionId).(*Metrics)
		if ok && metricsSystem._enabled {
			_ctx := context.Background()
			if instruments := metricsSystem._metrics.Get(ctx.system.Config.meriicsKey); instruments != nil {
				instruments.ClientSpawnCount.Add(_ctx, 1, metric.WithAttributes(metricsSystem.CommonLabels(ctx)...))
			}
		}
	}
}

// 日志接口
// Info ...
func (ctx *clientContext) Info(sfmt string, args ...interface{}) {
	ctx.system.logger.Info(fmt.Sprintf("[%s]", ctx.self.ToString()), sfmt, args...)
}

// Debug ...
func (ctx *clientContext) Debug(sfmt string, args ...interface{}) {
	ctx.system.logger.Debug(fmt.Sprintf("[%s]", ctx.self.ToString()), sfmt, args...)
}

// Error ...
func (ctx *clientContext) Error(sfmt string, args ...interface{}) {
	ctx.system.logger.Error(fmt.Sprintf("[%s]", ctx.self.ToString()), sfmt, args...)
}

// Warning ...
func (ctx *clientContext) Warning(sfmt string, args ...interface{}) {
	ctx.system.logger.Warning(fmt.Sprintf("[%s]", ctx.self.ToString()), sfmt, args...)
}

// Fatal ...
func (ctx *clientContext) Fatal(sfmt string, args ...interface{}) {
	ctx.system.logger.Fatal(fmt.Sprintf("[%s]", ctx.self.ToString()), sfmt, args...)
}

// Panic ...
func (ctx *clientContext) Panic(sfmt string, args ...interface{}) {
	ctx.system.logger.Panic(fmt.Sprintf("[%s]", ctx.self.ToString()), sfmt, args...)
}

func (ctx *clientContext) invokerAccept() {
	ctx.self.ref(ctx.system)
	ctx.client.Accept(ctx)
}

// MessageInvoker 接口实现
func (ctx *clientContext) invokerRecvice(b []byte, addr net.Addr) {

	defer func() {
		if r := recover(); r != nil {
			ctx.EscalateFailure(r, b)
		}
	}()

	if atomic.LoadInt32(&ctx.state) == stateClosed {
		// 已经关闭
		return
	}

	systemMetrics, ok := ctx.system.extensions.Get(ctx.system.extensionId).(*Metrics)
	if ok && systemMetrics._enabled {
		t := time.Now()
		ctx.message = b
		ctx.messageFrom = addr
		ctx.client.Recvice(ctx)
		ctx.messageFrom = nil
		ctx.message = nil

		delta := time.Since(t)
		context := context.Background()

		if instruments := systemMetrics._metrics.Get(ctx.system.Config.meriicsKey); instruments != nil {
			histogram := instruments.ClientBytesRecviceHistogram

			labels := append(
				systemMetrics.CommonLabels(ctx),
				attribute.String("message bytes", fmt.Sprintf("%d", len(b))),
			)
			histogram.Record(context, delta.Seconds(), metric.WithAttributes(labels...))
		}
	} else {
		ctx.message = b
		ctx.messageFrom = addr
		ctx.client.Recvice(ctx)
		ctx.messageFrom = nil
		ctx.message = nil
	}

}

func (ctx *clientContext) invokerPing() {
	if atomic.LoadInt32(&ctx.state) == stateClosed {
		// 已关闭
		return
	}

	ctx.client.Ping(ctx)
}

func (ctx *clientContext) invokerClosed() {
	if atomic.LoadInt32(&ctx.state) == stateClosed {
		// 已关闭
		return
	}

	atomic.StoreInt32(&ctx.state, stateClosed)
	atomic.StoreInt32(&ctx.self.vaild, 1)
	ctx.system.handlers.Remove(ctx.self)
	ctx.client.Closed(ctx)
}

func (ctx *clientContext) EscalateFailure(reason interface{}, message interface{}) {
}
