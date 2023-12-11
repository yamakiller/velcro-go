package network

import (
	"net"
	sync "sync"
)

func NewTcpNetworkServerModule(system *NetworkSystem) *tcpNetworkServerModule {
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
				break
			}
			conn, err = tns._listen.Accept()
			if err != nil {
				if e, ok := err.(net.Error); ok && e.Timeout() {
					continue
				}

				tns._system.Logger().Debug("[NETWORK-SERVER-TCP]", "%s accept error %s", address, err.Error())
				goto exit_label
			}

			//新建客户端
			if err = tns._system._props._producer(tns._system, conn); err != nil {
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
