package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	sync "sync"
	"time"

	"github.com/pkg/errors"
	"github.com/yamakiller/velcro-go/containers"
	"github.com/yamakiller/velcro-go/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func newTcpNetworkServerModule(system *NetworkSystem) *tcpNetworkServerModule {
	return &tcpNetworkServerModule{
		_system:    system,
		_waitGroup: sync.WaitGroup{},
		_stoped:    make(chan struct{}),
	}
}

type tcpNetworkServerModule struct {
	_system    *NetworkSystem
	_listen    net.Listener
	_waitGroup sync.WaitGroup
	_stoped    chan struct{}
}

func (tns *tcpNetworkServerModule) Open(addr string) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	tns._listen, err = net.ListenTCP("tcp", address)

	if err != nil {
		return err
	}

	tns._waitGroup.Add(1)

	go func() {
		defer tns._waitGroup.Done()

		var (
			err  error
			conn net.Conn
		)
		for {
			select {
			case <-tns._stoped:
				goto exit_label
			default:
			}
			conn, err = tns._listen.Accept()
			if err != nil {
				if e, ok := err.(net.Error); ok && e.Timeout() {
					continue
				}

				if strings.Contains(err.Error(), "use of closed network connection") {
					continue
				}

				tns._system.Debug("%s accept error %v", address, err)
				goto exit_label
			}

			//新建客户端
			if err = tns.spawn(conn); err != nil {
				conn.Close()
				continue
			}
		}
	exit_label:
	}()

	return nil
}

func (tns *tcpNetworkServerModule) Stop() {
	if tns._stoped != nil {
		close(tns._stoped)
	}

	if tns._listen != nil {
		tns._listen.Close()
	}

	tns._waitGroup.Wait()
	tns._stoped = nil
	tns._listen = nil
}

func (tns *tcpNetworkServerModule) spawn(conn net.Conn) error {
	id := tns._system._handlers.NextId()

	ctx := clientContext{_system: tns._system, _state: stateAccept}
	handler := &tcpClientHandler{
		_c:             conn,
		_wmail:         containers.NewQueue(4),
		_wmailcond:     *sync.NewCond(&sync.Mutex{}),
		_keepalive:     2000,
		_invoker:       &ctx,
		_senderStopped: make(chan struct{}),
		_closed:        0,
	}

	if tns._system.Config.MetricsProvider != nil {
		sysMetrics, ok := tns._system._extensions.Get(extensionId).(*Metrics)
		if ok && sysMetrics._enabled {
			if instruments := sysMetrics._metrics.Get(metrics.InternalClientMetrics); instruments != nil {
				sysMetrics.PrepareSendQueueLengthGauge()
				meter := otel.Meter(metrics.LibName)
				if _, err := meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveInt64(instruments.ClientSendQueueLength, int64(handler._wmail.Length()), metric.WithAttributes(sysMetrics.CommonLabels(&ctx)...))
					return nil
				}); err != nil {
					err = fmt.Errorf("failed to instrument Client SendQueue, %w", err)
					tns._system.Error(err.Error())
				}
			}
		}
	}

	cid, ok := tns._system._handlers.Push(handler, id)
	if !ok {
		return errors.Errorf("client-id %s existed", id)
	}

	ctx._self = cid
	ctx.incarnateClient()

	tns._waitGroup.Add(2)
	go func() {
		defer tns._waitGroup.Done()
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
		defer tns._waitGroup.Done()

		for {
			handler._wmailcond.L.Lock()
			if handler._started != 1 {
				handler._wmailcond.Wait()
			}
			handler._wmailcond.L.Unlock()

			if handler._started == 1 {
				break
			}
		}

		handler._invoker.invokerAccept()

		var tmp [512]byte
		remoteAddr := handler._c.RemoteAddr()
		for {

			if handler._keepalive > 0 {
				handler._c.SetReadDeadline(time.Now().Add(time.Duration(handler._keepalive) * time.Millisecond * 2.0))
			}

			n, err := handler._c.Read(tmp[:])
			if err != nil {
				if e, ok := err.(net.Error); ok && e.Timeout() {
					handler._keepaliveError++
					if handler._keepaliveError <= 3 {
						handler._invoker.invokerPing()
						continue
					}
				}
				break
			}

			handler._keepaliveError = 0
			handler._invoker.invokerRecvice(tmp[:n], &remoteAddr)
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
