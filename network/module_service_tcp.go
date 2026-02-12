package network

import (
	"context"
	"net"
	"strings"
	sync "sync"

	"github.com/pkg/errors"
	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/vlog"
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
		vlog.Infof("VELCRO: network server listen failed, addr=%s error=%s", addr, err)
		return err
	}
	vlog.Infof("VELCRO: network server listen at addr=%s", addr)

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

func (t *tcpNetworkServerModule) Stop() {
	if t.stoped != nil {
		close(t.stoped)
	}

	if t.listen != nil {
		t.listen.Close()
	}

	// t.waitGroup.Wait()
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
		sendbox:   circbuf.NewLinkBuffer(4096),
		sendcond:  sync.NewCond(&sync.Mutex{}),
		keepalive: uint32(t.system.Config.Kleepalive),
		invoker:   &ctx,
		// mailbox:   make(chan interface{}, 512),
		mailbox: spawnQueue(8),
		stopper: make(chan struct{}),
		refdone: &t.waitGroup,
	}

	cid, ok := t.system.handlers.Push(handler, id)
	if !ok {
		handler.Close()
		// 释放资源
		// close(handler.mailbox)
		close(handler.stopper)
		handler.sendbox.Close()
		handler.sendbox = nil
		handler.sendcond = nil

		return errors.Errorf("client-id %s existed", cid.ToString())
	}

	ctx.self = cid
	ctx.incarnateClient()

	handler.start()
	return nil
}
