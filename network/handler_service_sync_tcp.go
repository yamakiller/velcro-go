package network

import (
	"context"
	"errors"
	"net"
	sync "sync"
	"time"

	"github.com/yamakiller/velcro-go/gofunc"
)

type tcpSyncClientHandler struct {
	conn           net.Conn
	mutex          sync.Locker
	keepalive      uint32
	keepaliveError uint8
	invoker        MessageInvoker
	done           sync.WaitGroup
	refdone        *sync.WaitGroup
	stopper        chan struct{}
	ClientHandler
}

func (c *tcpSyncClientHandler) start() {
	c.refdone.Add(1)
	c.done.Add(1)

	gofunc.RecoverGoFuncWithInfo(context.Background(),
		c.reader,
		gofunc.NewBasicInfo("reader", c.invoker.invokerEscalateFailure))

}

// PostMessage 直接写数据
func (c *tcpSyncClientHandler) PostMessage(b []byte) error {
	if c.isStopped() {
		return errors.New("client: closed")
	}

	if _, err := c.conn.Write(b); err != nil {
		return err
	}

	return nil
}

func (c *tcpSyncClientHandler) PostToMessage(b []byte, target net.Addr) error {
	return errors.New("client: sync undefine post to message")
}

// Close 关闭连接
func (c *tcpSyncClientHandler) Close() {
	c.mutex.Lock()
	if !c.isStopped() {
		close(c.stopper)
	}
	c.mutex.Unlock()

	c.conn.Close()
	c.done.Wait()
}

func (c *tcpSyncClientHandler) isStopped() bool {
	select {
	case <-c.stopper:
		return true
	default:
		return false
	}
}

// reader 读取数据
func (c *tcpSyncClientHandler) reader() {
	defer c.done.Done()
	defer c.refdone.Done()

	var tmp [1024]byte
	remoteAddr := c.conn.RemoteAddr()

	c.invoker.invokerAccept()

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
					c.invoker.invokerPing()
					continue
				}
			}
			break
		}

		c.keepaliveError = 0
		c.invoker.invokerRecvice(tmp[:n], remoteAddr)
	}

	c.conn.Close()
	c.mutex.Lock()
	if !c.isStopped() {
		close(c.stopper)
	}
	c.mutex.Unlock()

	c.invoker.invokerClosed()
}
