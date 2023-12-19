package network

import (
	"net"
	sync "sync"

	"github.com/pkg/errors"
	"github.com/yamakiller/velcro-go/containers"
	"github.com/yamakiller/velcro-go/syncx"
)

func newUDPNetworkServerModule(system *NetworkSystem) *udpNetworkServerModule {
	return &udpNetworkServerModule{
		_system:    system,
		_waitGroup: sync.WaitGroup{},
		_stoped:    make(chan struct{}),
	}
}

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
		conn:     uns._listen,
		sendbox:  containers.NewQueue(4, &syncx.NoMutex{}),
		sendcond: sync.NewCond(&sync.Mutex{}),
		invoker:  &ctx,
		mailbox:  make(chan interface{}, 1),
		stopper:  make(chan struct{}),
		refdone:  &uns._waitGroup,
	}

	cid, ok := uns._system._handlers.Push(handler, id)
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

	ctx._self = cid
	ctx.incarnateClient()

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
