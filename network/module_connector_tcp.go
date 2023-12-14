package network

import (
	"context"
	"fmt"
	"net"
	sync "sync"
	"time"

	"github.com/pkg/errors"
	"github.com/yamakiller/velcro-go/containers"
	"github.com/yamakiller/velcro-go/metrics"
	lsync "github.com/yamakiller/velcro-go/sync"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func newTcpConnectorNetworkServerModule(system *NetworkSystem) *tcpNetworkConnectorModule {
	return &tcpNetworkConnectorModule{
		_system:    system,
		_waitGroup: sync.WaitGroup{},
		_stoped:    make(chan struct{}),
	}
}

type tcpNetworkConnectorModule struct {
	_system    *NetworkSystem
	_waitGroup sync.WaitGroup
	_stoped    chan struct{}
}

func (tnc *tcpNetworkConnectorModule) Open(addr string) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp",
		address.String(),
		time.Duration(tnc._system.Config.NetowkTimeout)*time.Millisecond)
	if err != nil {
		return err
	}

	if err = tnc.spawn(conn); err != nil {
		conn.Close()
		return err
	}

	return nil
}

func (tnc *tcpNetworkConnectorModule) Network() string {
	return "tcpconnector"
}

func (tnc *tcpNetworkConnectorModule) Stop() {
	if tnc._stoped != nil {
		close(tnc._stoped)
	}

	tnc._waitGroup.Wait()
	tnc._stoped = nil
}

func (tnc *tcpNetworkConnectorModule) isStopped() bool {
	select {
	case <-tnc._stoped:
		return true
	default:
		return false
	}
}

// spawn 创建TCP Connector 对象
func (tnc *tcpNetworkConnectorModule) spawn(conn net.Conn) error {
	id := tnc._system._handlers.NextId()

	ctx := clientContext{_system: tnc._system, _state: stateAccept}
	handler := &tcpConnectorHandler{
		_c:             conn,
		_wmail:         containers.NewQueue(4, &lsync.NoMutex{}),
		_wmailcond:     sync.NewCond(&sync.Mutex{}),
		_invoker:       &ctx,
		_senderStopped: make(chan struct{}),
	}

	if tnc._system.Config.MetricsProvider != nil {
		sysMetrics, ok := tnc._system._extensions.Get(ctx._system._extensionId).(*Metrics)
		if ok && sysMetrics._enabled {
			if instruments := sysMetrics._metrics.Get(ctx._system.Config.meriicsKey); instruments != nil {
				sysMetrics.PrepareSendQueueLengthGauge()
				meter := otel.Meter(metrics.LibName)
				if _, err := meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveInt64(instruments.ClientSendQueueLength, int64(handler._wmail.Length()), metric.WithAttributes(sysMetrics.CommonLabels(&ctx)...))
					return nil
				}); err != nil {
					err = fmt.Errorf("failed to instrument Client SendQueue, %w", err)
					tnc._system.Error(err.Error())
				}
			}
		}
	}

	cid, ok := tnc._system._handlers.Push(handler, id)
	if !ok {
		handler._c.Close()
		handler._wmail = nil
		handler._wmailcond = nil
		close(handler._senderStopped)
		return errors.Errorf("client-id %s existed", id)
	}

	ctx._self = cid
	ctx.incarnateClient()

	tnc._waitGroup.Add(2)

	go func() {
		defer tnc._waitGroup.Done()
		for {
			handler._wmailcond.L.Lock()
			msg, ok := handler._wmail.Pop()
			if !ok || (handler._started == 0 && msg != nil) {
				handler._wmailcond.Wait()
			}

			if handler._closed != 0 {
				handler._wmailcond.L.Unlock()
				break
			}
			handler._wmailcond.L.Unlock()

			if msg == nil && ok {
				break
			} else if msg != nil {
				if _, err := handler._c.Write(msg.([]byte)); err != nil {
					break
				}
			}
		}

		handler._wmailcond.L.Lock()
		handler._closed = 1
		handler._wmailcond.L.Unlock()

		handler._c.Close()
		handler._wmail.Destory()
		handler._wmail = nil
		close(handler._senderStopped)
	}()

	go func() {
		defer tnc._waitGroup.Done()

		for {
			handler._wmailcond.L.Lock()
			if handler._started != 1 && !tnc.isStopped() {
				handler._wmailcond.Wait()
			}
			handler._wmailcond.L.Unlock()

			if handler._started == 1 || tnc.isStopped() {
				break
			}
		}

		handler._invoker.invokerAccept()

		var tmp [512]byte
		remoteAddr := handler._c.RemoteAddr()
		for {

			if tnc.isStopped() || handler._closed != 0 {
				break
			}

			// 防止挂死
			handler._c.SetReadDeadline(time.Now().Add(time.Duration(tnc._system.Config.NetowkTimeout) * time.Millisecond * 2.0))

			n, err := handler._c.Read(tmp[:])
			if err != nil {
				if e, ok := err.(net.Error); ok && e.Timeout() {
					continue
				}
				break
			}

			handler._invoker.invokerRecvice(tmp[:n], remoteAddr)
		}

		handler._wmailcond.L.Lock()
		if handler._closed == 0 {
			handler._closed = 1
			handler._c.Close()
		}
		handler._wmailcond.L.Unlock()
		handler._wmailcond.Signal()

		// 等待发送端结束
		_ = <-handler._senderStopped

		//调用已关闭
		if handler._invoker != nil {
			handler._invoker.invokerClosed()
		}

	}()

	handler.start()

	return nil
}
