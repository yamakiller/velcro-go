package network

import (
	"net"
	sync "sync"

	"github.com/yamakiller/velcro-go/containers"
)

type udpMsg struct {
	data []byte
	addr *net.UDPAddr
}

type udpClientHandler struct {
	conn      *net.UDPConn
	sendbox   *containers.Queue
	sendcond  *sync.Cond
	mailbox   chan interface{}
	invoker   MessageInvoker
	done      sync.WaitGroup
	guarddone sync.WaitGroup
	refdone   *sync.WaitGroup
	stopper   chan struct{}
	ClientHandler
}

func (u *udpClientHandler) start() {
	u.refdone.Add(3)
	u.done.Add(2)
	go u.sender()
	go u.reader()
	u.guarddone.Add(1)
	go u.guardian()
}

func (u *udpClientHandler) PostMessage(b []byte) {
	panic("udp undefine post message")
}

func (u *udpClientHandler) PostToMessage(b []byte, target net.Addr) {

	udpTarget, ok := target.(*net.UDPAddr)
	if !ok {
		panic("udp post to message Please enter net.UDPAddr")
	}

	u.sendcond.L.Lock()
	if u.isStopped() {
		u.sendcond.L.Unlock()
		return
	}

	u.sendbox.Push(&udpMsg{data: b, addr: udpTarget})
	u.sendcond.L.Unlock()
	u.sendcond.Signal()
}

func (u *udpClientHandler) Close() {
	u.sendcond.L.Lock()
	if u.isStopped() {
		u.sendcond.L.Unlock()
		return
	}

	u.sendbox.Push(nil)
	u.sendcond.L.Unlock()
	u.sendcond.Signal()

	u.done.Wait()
	u.guarddone.Wait()
}

func (u *udpClientHandler) isStopped() bool {
	select {
	case <-u.stopper:
		return true
	default:
		return false
	}
}

func (u *udpClientHandler) sender() {
	defer u.done.Done()
	defer u.refdone.Done()

	for {
		u.sendcond.L.Lock()
		if !u.isStopped() {
			u.sendcond.Wait()
		}
		u.sendcond.L.Unlock()

		for {
			if u.isStopped() {
				goto udp_sender_exit_label
			}

			u.sendcond.L.Lock()
			msg, ok := u.sendbox.Pop()
			u.sendcond.L.Unlock()

			if !ok {
				break
			} else if msg == nil && ok {
				goto udp_sender_exit_label
			}

			if msg != nil {

				udpD := msg.(*udpMsg)
				if _, err := u.conn.WriteToUDP(udpD.data, udpD.addr); err != nil {
					goto udp_sender_exit_label
				}
			}
		}
	}
udp_sender_exit_label:
	close(u.stopper)
	u.conn.Close()
}

func (u *udpClientHandler) reader() {
	defer u.done.Done()
	defer u.refdone.Done()

	u.mailbox <- &AcceptMessage{}

	var tmp [512]byte

	for {

		if u.isStopped() {
			break
		}

		n, addr, err := u.conn.ReadFromUDP(tmp[:])
		if err != nil {
			break
		}

		u.mailbox <- &RecviceMessage{Data: tmp[:n], Addr: addr}
	}

	u.conn.Close()
	u.sendcond.Signal()
	u.mailbox <- &ClosedMessage{}
}

func (u *udpClientHandler) guardian() {
	defer u.guarddone.Done()
	defer u.refdone.Done()

	for {
		msg, ok := <-u.mailbox
		if !ok {
			goto udp_guardian_exit_lable
		}

		switch message := msg.(type) {
		case *AcceptMessage:
			u.invoker.invokerAccept()
		case *RecviceMessage:
			u.invoker.invokerRecvice(message.Data, message.Addr)
		case *PingMessage:
			u.invoker.invokerPing()
		case *ClosedMessage:
			goto udp_guardian_exit_lable
		default:
			panic("udp client guardian: unknown message")
		}
	}
udp_guardian_exit_lable:
	close(u.mailbox)
	u.done.Wait()

	// 释放资源
	u.sendbox.Destory()
	u.sendbox = nil
	u.sendcond = nil

	u.invoker.invokerClosed()
}
