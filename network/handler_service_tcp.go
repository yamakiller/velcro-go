package network

import (
	"context"
	"errors"
	"net"
	"runtime"
	sync "sync"
	"time"

	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/vlog"
)

var _ Handler = &tcpClientHandler{}

// ClientHandler TCP服务客户端处理程序
type tcpClientHandler struct {
	conn     net.Conn
	sendbox  *circbuf.LinkBuffer
	sendcond *sync.Cond
	// mailbox  chan interface{}
	mailbox        *queue
	keepalive      uint32
	keepaliveError uint8
	invoker        MessageInvoker
	done           sync.WaitGroup
	guarddone      sync.WaitGroup
	refdone        *sync.WaitGroup
	stopper        chan struct{}
	ClientHandler
}

func (c *tcpClientHandler) start() {
	c.refdone.Add(3)
	c.done.Add(2)

	gofunc.RecoverGoFuncWithInfo(context.Background(),
		c.sender,
		gofunc.NewBasicInfo("sender", c.invoker.invokerEscalateFailure))

	gofunc.RecoverGoFuncWithInfo(context.Background(),
		c.reader,
		gofunc.NewBasicInfo("reader", c.invoker.invokerEscalateFailure))

	c.guarddone.Add(1)

	gofunc.RecoverGoFuncWithInfo(context.Background(),
		c.guardian,
		gofunc.NewBasicInfo("guardian", c.invoker.invokerEscalateFailure))
}

func (c *tcpClientHandler) PostMessage(b []byte) error {
	c.sendcond.L.Lock()
	if c.isStopped() {
		c.sendcond.L.Unlock()
		return errors.New("client: closed")
	}
	if _, err := c.sendbox.WriteBinary(b); err != nil {
		c.sendcond.L.Unlock()
		return err
	}

	c.sendcond.Signal()
	c.sendcond.L.Unlock()
	return nil
}

func (c *tcpClientHandler) PostToMessage(b []byte, target net.Addr) error {
	// c.sendcond.L.Lock()
	// if c.isStopped() {
	// 	c.sendcond.L.Unlock()
	// 	return errors.New("client: closed")
	// }
	// c.sendcond.Signal()
	// c.sendcond.L.Unlock()
	rec := &RecviceMessage{
		Data: make([]byte, len(b)),
		Addr: target,
	}
	copy(rec.Data, b)
	c.mailbox.Push(rec)
	// c.mailbox <- rec
	return nil
}

func (c *tcpClientHandler) Close() {
	c.sendcond.L.Lock()
	if c.isStopped() {
		c.sendcond.L.Unlock()
		return
	}

	c.conn.Close()
	close(c.stopper)

	c.sendcond.Signal()
	c.sendcond.L.Unlock()

}

func (c *tcpClientHandler) isStopped() bool {
	select {
	case <-c.stopper:
		return true
	default:
		return false
	}
}

func (c *tcpClientHandler) sender() {
	defer func() {
		c.done.Done()
		c.refdone.Done()
	}()

	var (
		err       error
		readbytes []byte = nil
	)
	for {
		c.sendcond.L.Lock()
		if !c.isStopped() && c.sendbox.MallocLen() == 0 {
			c.sendcond.Wait()
		}
		c.sendcond.L.Unlock()

		for {
			if c.isStopped() {
				goto tcp_sender_exit_label
			}

			c.sendcond.L.Lock()
			if c.sendbox.MallocLen() != 0 {
				c.sendbox.Flush()
			}
			if c.sendbox.Len() > 0 {
				readbytes, err = c.sendbox.ReadBinary(c.sendbox.Len())
				if err != nil {
					c.sendcond.L.Unlock()
					vlog.Errorf("tcp handler error sendbuffer readbinary fail %s", err.Error())
					goto tcp_sender_exit_label
				}
			}
			if readbytes == nil {
				c.sendcond.L.Unlock()
				break
			}
			c.sendcond.L.Unlock()

			i := 0
			offset := 0
			nwrite := 0
			for {

				if i > 1 {
					runtime.Gosched()
					i = 0
				}

				if c.isStopped() {
					goto tcp_sender_exit_label
				}

				// c.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 50))
				if nwrite, err = c.conn.Write(readbytes[offset:]); err != nil {
					// if e, ok := err.(net.Error); ok && e.Timeout() {
					// 	goto tcp_sender_continue_label
					// }

					goto tcp_sender_exit_label
				}
				// tcp_sender_continue_label:
				offset += nwrite
				if offset == len(readbytes) {
					break
				}
				i++
			}

			readbytes = nil

		}
	}
tcp_sender_exit_label:
	c.sendcond.L.Lock()
	if !c.isStopped() {
		close(c.stopper)
	}
	c.sendcond.L.Unlock()
	c.conn.Close()
}

func (c *tcpClientHandler) reader() {
	defer func() {
		c.done.Done()
		c.refdone.Done()
	}()
	// c.mailbox <- &AcceptMessage{}
	c.mailbox.Push(&AcceptMessage{})
	var tmp [512]byte
	remoteAddr := c.conn.RemoteAddr()
	for {

		if c.isStopped() {
			break
		}
		if c.keepalive > 0 {
			c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.keepalive) * time.Millisecond * 2.0))
		}

		n, err := c.conn.Read(tmp[:])
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				c.keepaliveError++
				if c.keepaliveError <= 3 {
					// c.mailbox <- &PingMessage{}
					c.mailbox.Push(&PingMessage{})
					continue
				}
			}
			break
		}
		c.keepaliveError = 0
		rec := &RecviceMessage{
			Data: make([]byte, n),
			Addr: remoteAddr,
		}
		copy(rec.Data, tmp[:n])
		// c.mailbox <- rec
		c.mailbox.Push(rec)
	}

	c.conn.Close()
	c.sendcond.L.Lock()
	if !c.isStopped() {
		close(c.stopper)
	}
	c.sendcond.Signal()
	c.sendcond.L.Unlock()

	// c.mailbox <- &ClosedMessage{}
	c.mailbox.Push(&ClosedMessage{})
}

func (c *tcpClientHandler) guardian() {

	defer func() {
		c.guarddone.Done()
		c.refdone.Done()
	}()
	for {
		// msg, ok := <-c.mailbox
		msg, ok := c.mailbox.Pop()
		if !ok {
			continue
			// goto tcp_guardian_exit_lable
		}

		switch message := msg.(type) {
		case *AcceptMessage:
			c.invoker.invokerAccept()
		case *RecviceMessage:
			c.invoker.invokerRecvice(message.Data, message.Addr)
		case *PingMessage:
			c.invoker.invokerPing()
		case *ClosedMessage:
			goto tcp_guardian_exit_lable
		default:
			panic("tcp client guardian: unknown message")
		}
	}
tcp_guardian_exit_lable:
	// close(c.mailbox)
	c.done.Wait()

	// 释放资源
	c.sendbox.Close()
	// c.sendbox = nil
	// c.sendcond = nil

	c.invoker.invokerClosed()
}
