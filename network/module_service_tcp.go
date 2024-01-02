package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	sync "sync"

	"github.com/pkg/errors"
	"github.com/yamakiller/velcro-go/containers"
	"github.com/yamakiller/velcro-go/debugs/metrics"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func newTCPNetworkServerModule(system *NetworkSystem) *tcpNetworkServerModule {
	return &tcpNetworkServerModule{
		system:    system,
		waitGroup: sync.WaitGroup{},
		stoped:    make(chan struct{}),
	}
}

type tcpNetworkServerModule struct {
	system    *NetworkSystem
	listen    net.Listener
	waitGroup sync.WaitGroup
	stoped    chan struct{}
}

func (t *tcpNetworkServerModule) Open(addr string) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	t.listen, err = net.ListenTCP("tcp", address)

	if err != nil {
		return err
	}

	t.waitGroup.Add(1)

	go func() {
		defer t.waitGroup.Done()

		var (
			err  error
			conn net.Conn
		)
		for {
			select {
			case <-t.stoped:
				goto exit_label
			default:
			}
			conn, err = t.listen.Accept()
			if err != nil {
				if e, ok := err.(net.Error); ok && e.Timeout() {
					continue
				}

				if strings.Contains(err.Error(), "use of closed network connection") {
					continue
				}

				t.system.Debug("%s accept error %v", address, err)
				goto exit_label
			}

			//新建客户端
			if err = t.spawn(conn); err != nil {
				conn.Close()
				continue
			}
		}
	exit_label:
	}()

	return nil
}

func (t *tcpNetworkServerModule) Stop() {
	if t.stoped != nil {
		close(t.stoped)
	}

	if t.listen != nil {
		t.listen.Close()
	}

	t.waitGroup.Wait()
	t.stoped = nil
	t.listen = nil
}

func (tnc *tcpNetworkServerModule) Network() string {
	return "tcpserver"
}

func (t *tcpNetworkServerModule) isStopped() bool {
	select {
	case <-t.stoped:
		return true
	default:
		return false
	}
}

func (t *tcpNetworkServerModule) spawn(conn net.Conn) error {
	id := t.system.handlers.NextId()

	ctx := clientContext{system: t.system, state: stateAccept}
	handler := &tcpClientHandler{
		conn:      conn,
		sendbox:   containers.NewQueue(4, &syncx.NoMutex{}),
		sendcond:  sync.NewCond(&sync.Mutex{}),
		keepalive: uint32(t.system.Config.Kleepalive),
		invoker:   &ctx,
		mailbox:   make(chan interface{}, 1),
		stopper:   make(chan struct{}),
		refdone:   &t.waitGroup,
	}

	if t.system.Config.MetricsProvider != nil {
		sysMetrics, ok := t.system.extensions.Get(t.system.extensionId).(*Metrics)
		if ok && sysMetrics._enabled {
			if instruments := sysMetrics._metrics.Get(ctx.system.Config.meriicsKey); instruments != nil {
				sysMetrics.PrepareSendQueueLengthGauge()
				meter := otel.Meter(metrics.LibName)
				if _, err := meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveInt64(instruments.ClientSendQueueLength, int64(handler.sendbox.Length()), metric.WithAttributes(sysMetrics.CommonLabels(&ctx)...))
					return nil
				}); err != nil {
					err = fmt.Errorf("failed to instrument Client SendQueue, %w", err)
					t.system.Error(err.Error())
				}
			}
		}
	}

	cid, ok := t.system.handlers.Push(handler, id)
	if !ok {
		handler.Close()
		// 释放资源
		close(handler.mailbox)
		close(handler.stopper)
		handler.sendbox.Destory()
		handler.sendbox = nil
		handler.sendcond = nil

		return errors.Errorf("client-id %s existed", cid.ToString())
	}

	ctx.self = cid
	ctx.incarnateClient()

	handler.start()
	return nil
}
