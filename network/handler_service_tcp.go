package network

import (
	"context"
	"errors"
	"net"
	sync "sync"
	"time"

	"github.com/yamakiller/velcro-go/gofunc"
)

var _ Handler = &tcpClientHandler{}

// ClientHandler TCP服务客户端处理程序
type tcpClientHandler struct {
	conn           net.Conn
	// sendbox        *circbuf.LinkBuffer
	sendcond       *sync.Cond
	mailbox        chan interface{}
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
	c.done.Add(1)

	// gofunc.RecoverGoFuncWithInfo(context.Background(),
	// 	c.sender,
	// 	gofunc.NewBasicInfo("sender", c.invoker.invokerEscalateFailure))

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
	c.conn.Write(b)
	// c.sendbox.WriteBinary(b)
	c.sendcond.L.Unlock()

	c.sendcond.Signal()

	return nil
}

func (c *tcpClientHandler) PostToMessage(b []byte, target net.Addr) error {
	return errors.New("client: undefine post to message")
}

func (c *tcpClientHandler) Close() {
	c.sendcond.L.Lock()
	if c.isStopped() {
		c.sendcond.L.Unlock()
		return
	}

	c.conn.Close()

	c.sendcond.L.Unlock()
	c.sendcond.Signal()

	c.done.Wait()
	c.guarddone.Wait()
}

func (c *tcpClientHandler) isStopped() bool {
	select {
	case <-c.stopper:
		return true
	default:
		return false
	}
}

// func (c *tcpClientHandler) sender() {
// 	defer func ()  {
// 		c.done.Done()
// 		c.refdone.Done()
// 	}()
 
// 	var (
// 		err       error
// 		readbytes []byte = nil
// 	)
// 	for {
// 		c.sendcond.L.Lock()
// 		if !c.isStopped() {
// 			c.sendcond.Wait()
// 		}
// 		c.sendcond.L.Unlock()
// 		i := 0
// 		for {
// 			if c.isStopped() {
// 				goto tcp_sender_exit_label
// 			}

// 			c.sendcond.L.Lock()
// 			if c.sendbox.Len() > 0 {
// 				readbytes, err = c.sendbox.ReadBinary(c.sendbox.Len())
// 				if err != nil {
// 					c.sendcond.L.Unlock()
// 					vlog.Errorf("tcp handler error sendbuffer readbinary fail %s", err.Error())
// 					goto tcp_sender_exit_label
// 				}
// 			}
// 			c.sendcond.L.Unlock()

// 			if readbytes != nil {
// 				if i > 1 {
// 					runtime.Gosched()
// 					i = 0
// 				}

// 				if _, err := c.conn.Write(readbytes); err != nil {
// 					goto tcp_sender_exit_label
// 				}
// 				readbytes = nil
// 				i++
// 			}else{
// 				break
// 			}
// 		}
// 	}
// tcp_sender_exit_label:
// 	close(c.stopper)
// 	c.conn.Close()
// }

func (c *tcpClientHandler) reader() {
	defer func ()  {
		c.done.Done()
		c.refdone.Done()
	}()
	c.mailbox <- &AcceptMessage{}

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
					c.mailbox <- &PingMessage{}
					continue
				}
			}
			break
		}

		c.keepaliveError = 0
		c.mailbox <- &RecviceMessage{Data: tmp[:n], Addr: remoteAddr}
	}

	c.conn.Close()
	c.sendcond.Signal()

	c.mailbox <- &ClosedMessage{}

}

func (c *tcpClientHandler) guardian() {
	defer c.guarddone.Done()
	defer c.refdone.Done()

	for {
		msg, ok := <-c.mailbox
		if !ok {
			goto tcp_guardian_exit_lable
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
	close(c.mailbox)
	c.done.Wait()

	// 释放资源
	// c.sendbox.Close()
	// c.sendbox = nil
	// c.sendcond = nil

	c.invoker.invokerClosed()
}
