package network

import (
	"net"
	sync "sync"

	"github.com/pkg/errors"
	"github.com/yamakiller/velcro-go/utils/collection"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"github.com/yamakiller/velcro-go/vlog"
)

func newUDPNetworkServerModule(system *NetworkSystem) *udpNetworkServerModule {
	return &udpNetworkServerModule{
		system:    system,
		waitGroup: sync.WaitGroup{},
		stoped:    make(chan struct{}),
	}
}

type udpNetworkServerModule struct {
	system    *NetworkSystem
	listen    *net.UDPConn
	waitGroup sync.WaitGroup
	stoped    chan struct{}
}

func (u *udpNetworkServerModule) Open(addr string) error {
	address, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	u.listen, err = net.ListenUDP("udp", address)
	if err != nil {
		vlog.Infof("VELCRO: network server udp listen failed, addr=%s error=%s", addr, err)
		return err
	}
	vlog.Errorf("VELCRO: network server udp listen at addr=%s", addr)

	id := u.system.handlers.NextId()

	ctx := clientContext{system: u.system, state: stateAccept}
	handler := &udpClientHandler{
		conn:     u.listen,
		sendbox:  collection.NewQueue(4, &syncx.NoMutex{}),
		sendcond: sync.NewCond(&sync.Mutex{}),
		invoker:  &ctx,
		mailbox:  make(chan interface{}, 1),
		stopper:  make(chan struct{}),
		refdone:  &u.waitGroup,
	}

	cid, ok := u.system.handlers.Push(handler, id)
	if !ok {
		handler.Close()
		// 清理资源
		close(handler.mailbox)
		close(handler.stopper)
		handler.sendbox.Destory()
		handler.sendbox = nil
		handler.sendcond = nil

		return errors.Errorf("client-id %s existed", id)
	}

	ctx.self = cid
	ctx.incarnateClient()

	handler.start()
	return nil
}

func (u *udpNetworkServerModule) Stop() {
	if u.stoped != nil {
		close(u.stoped)
	}

	if u.listen != nil {
		u.listen.Close()
	}

	u.waitGroup.Wait()
	u.stoped = nil
	u.listen = nil
}

func (u *udpNetworkServerModule) Network() string {
	return "udpserver"
}

func (u *udpNetworkServerModule) isStopped() bool {
	select {
	case <-u.stoped:
		return true
	default:
		return false
	}
}
