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
	lsync "github.com/yamakiller/velcro-go/sync"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func newTCPNetworkServerModule(system *NetworkSystem) *tcpNetworkServerModule {
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

func (tnc *tcpNetworkServerModule) Network() string {
	return "tcpserver"
}

func (tns *tcpNetworkServerModule) isStopped() bool {
	select {
	case <-tns._stoped:
		return true
	default:
		return false
	}
}

func (tns *tcpNetworkServerModule) spawn(conn net.Conn) error {
	id := tns._system._handlers.NextId()

	ctx := clientContext{_system: tns._system, _state: stateAccept}
	handler := &tcpClientHandler{
		conn:      conn,
		sendbox:   containers.NewQueue(4, &lsync.NoMutex{}),
		sendcond:  sync.NewCond(&sync.Mutex{}),
		keepalive: uint32(tns._system.Config.NetowkTimeout),
		invoker:   &ctx,
		mailbox:   make(chan interface{}, 1),
		stopper:   make(chan struct{}),
		refdone:   &tns._waitGroup,
	}

	if tns._system.Config.MetricsProvider != nil {
		sysMetrics, ok := tns._system._extensions.Get(tns._system._extensionId).(*Metrics)
		if ok && sysMetrics._enabled {
			if instruments := sysMetrics._metrics.Get(ctx._system.Config.meriicsKey); instruments != nil {
				sysMetrics.PrepareSendQueueLengthGauge()
				meter := otel.Meter(metrics.LibName)
				if _, err := meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveInt64(instruments.ClientSendQueueLength, int64(handler.sendbox.Length()), metric.WithAttributes(sysMetrics.CommonLabels(&ctx)...))
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
		handler.Close()
		// 释放资源
		close(handler.mailbox)
		close(handler.stopper)
		handler.sendbox.Destory()
		handler.sendbox = nil
		handler.sendcond = nil

		return errors.Errorf("client-id %s existed", cid.ToString())
	}

	ctx._self = cid
	ctx.incarnateClient()

	handler.start()
	return nil
}
