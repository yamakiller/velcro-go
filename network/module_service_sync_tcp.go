package network

import (
	"context"
	"net"
	"strings"
	sync "sync"

	"github.com/pkg/errors"
	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/vlog"
)

type tcpSyncNetworkServerModule struct {
	system    *NetworkSystem
	listen    net.Listener
	waitGroup sync.WaitGroup
	stoped    chan struct{}
}

func (t *tcpSyncNetworkServerModule) Open(addr string) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	t.listen, err = net.ListenTCP("tcp", address)

	if err != nil {
		vlog.Infof("VELCRO: network server listen failed, addr=%s error=%s", addr, err)
		return err
	}
	vlog.Errorf("VELCRO: network server listen at addr=%s", addr)

	t.waitGroup.Add(1)

	gofunc.GoFunc(context.Background(), func() {
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

				vlog.Debugf("%s accept error %v", address, err)
				goto exit_label
			}

			//新建客户端
			if err = t.spawn(conn); err != nil {
				conn.Close()
				continue
			}
		}
	exit_label:
	})

	return nil
}

func (t *tcpSyncNetworkServerModule) Stop() {
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

func (t *tcpSyncNetworkServerModule) Network() string {
	return "tcp-sync-server"
}

func (t *tcpSyncNetworkServerModule) isStopped() bool {
	select {
	case <-t.stoped:
		return true
	default:
		return false
	}
}

func (t *tcpSyncNetworkServerModule) spawn(conn net.Conn) error {
	id := t.system.handlers.NextId()

	ctx := clientContext{system: t.system, state: stateAccept}
	handler := &tcpSyncClientHandler{
		conn:      conn,
		mutex:     &sync.Mutex{},
		keepalive: uint32(t.system.Config.Kleepalive),
		invoker:   &ctx,
		stopper:   make(chan struct{}),
		refdone:   &t.waitGroup,
	}

	cid, ok := t.system.handlers.Push(handler, id)
	if !ok {
		handler.Close()
		close(handler.stopper)
		return errors.Errorf("client-id %s existed", cid.ToString())
	}

	ctx.self = cid
	ctx.incarnateClient()

	handler.start()

	return nil
}
