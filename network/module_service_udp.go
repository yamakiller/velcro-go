package network

import (
	"context"
	"fmt"
	"net"
	sync "sync"
	"time"

	"github.com/pkg/errors"
	"github.com/yamakiller/velcro-go/containers"
	lsync "github.com/yamakiller/velcro-go/sync"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type udpNetworkServerModule struct {
	_system    *NetworkSystem
	_listen    *net.UDPConn
	_waitGroup sync.WaitGroup
	_stoped    chan struct{}
}

func (uns *udpNetworkServerModule) Open(addr string) error {
	address, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	uns._listen, err = net.ListenUDP("udp", address)
	if err != nil {
		return err
	}

	id := uns._system._handlers.NextId()

	ctx := clientContext{_system: uns._system, _state: stateAccept}
	handler := &udpClientHandler{
		_c:             uns._listen,
		_wmail:         containers.NewQueue(4, &lsync.NoMutex{}),
		_wmailcond:     sync.NewCond(&sync.Mutex{}),
		_invoker:       &ctx,
		_senderStopped: make(chan struct{}),
	}

	cid, ok := uns._system._handlers.Push(handler, id)
	if !ok {
		uns._listen.Close()
		handler._c = nil
		handler._wmailcond = nil
		handler._invoker = nil
		handler._wmail = nil
		close(handler._senderStopped)
		return errors.Errorf("client-id %s existed", id)
	}

	ctx._self = cid
	ctx.incarnateClient()

	uns._waitGroup.Add(2)

	go func() {
		defer uns._waitGroup.Done()
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
				systemMetrics, ok := ctx._system._extensions.Get(ctx._system._extensionId).(*Metrics)
				if ok && systemMetrics._enabled {
					t := time.Now()
					udpD := msg.(*udpMsg)
					if _, err := handler._c.WriteToUDP(udpD.data, udpD.addr); err != nil {
						break
					}

					delta := time.Since(t)
					_ctx := context.Background()

					if instruments := systemMetrics._metrics.Get(ctx._system.Config.meriicsKey); instruments != nil {
						histogram := instruments.ClientBytesSendHistogram

						labels := append(
							systemMetrics.CommonLabels(&ctx),
							attribute.String("message bytes", fmt.Sprintf("%d", len(msg.([]byte)))),
						)
						histogram.Record(_ctx, delta.Seconds(), metric.WithAttributes(labels...))
					}
				} else {
					udpD := msg.(*udpMsg)
					if _, err := handler._c.WriteToUDP(udpD.data, udpD.addr); err != nil {
						break
					}
				}

			}
		}

	}()

	go func() {
		defer uns._waitGroup.Done()

		for {
			handler._wmailcond.L.Lock()
			if handler._started != 1 && !uns.isStopped() {
				handler._wmailcond.Wait()
			}
			handler._wmailcond.L.Unlock()

			if handler._started == 1 || uns.isStopped() {
				break
			}
		}

		handler._invoker.invokerAccept()

		var tmp [1500]byte

		for {

			if uns.isStopped() || handler._closed != 0 {
				break
			}

			n, addr, err := handler._c.ReadFromUDP(tmp[:])
			if err != nil {
				break
			}

			handler._invoker.invokerRecvice(tmp[:n], addr)
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

func (uns *udpNetworkServerModule) Stop() {
	if uns._stoped != nil {
		close(uns._stoped)
	}

	if uns._listen != nil {
		uns._listen.Close()
	}

	uns._waitGroup.Wait()
	uns._stoped = nil
	uns._listen = nil
}

func (uns *udpNetworkServerModule) Network() string {
	return "udpserver"
}

func (uns *udpNetworkServerModule) isStopped() bool {
	select {
	case <-uns._stoped:
		return true
	default:
		return false
	}
}
