package tcpclient

import (
	"context"
	"fmt"
	// "crypto"
	// "encoding/base64"
	"errors"
	"net"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	"github.com/yamakiller/velcro-go/rpc/client"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Closed() {

}

type AcceptMessage struct {
}

type PingMessage struct {
}

type RecviceMessage struct {
	Data []byte
	Addr net.Addr
}

type ClosedMessage struct {
}

func NewConn(options ...client.ConnConfigOption) client.IConnect {
	config := client.Configure(options...)

	return NewConnConfig(config)
}

func NewConnConfig(config *client.ConnConfig) *Conn {
	return &Conn{
		BaseConnect: &client.BaseConnect{},
		Config:      config,
		sendbox:     circbuf.NewLinkBuffer(4096),
		recvice:     circbuf.NewLinkBuffer(4096),
		sendcond:    sync.NewCond(&sync.Mutex{}),
		waitbox:     cmap.New(),
		methods:     make(map[interface{}]func(ctx context.Context, message proto.Message) (proto.Message, error)),
		stopper:     make(chan struct{}),
		sequence:    1,
		mailbox:     make(chan interface{}, 1),
		state:       Disconnected,
	}
}

type ConnState int

const (
	Disconnected = iota
	Connecting
	Connected
	Disconnecting
)

type IntervalLinkNode struct {
	intrusive.LinkedNode
	Value interface{}
}

type Conn struct {
	*client.BaseConnect
	Config   *client.ConnConfig
	conn     net.Conn
	address  *net.TCPAddr
	timeout  time.Duration
	sendbox  *circbuf.LinkBuffer
	waitbox  cmap.ConcurrentMap
	sendcond *sync.Cond
	done     sync.WaitGroup

	recvice *circbuf.LinkBuffer

	stopper  chan struct{}
	sequence int32

	mailbox     chan interface{}
	mailboxDone sync.WaitGroup

	methods map[interface{}]func(ctx context.Context, message proto.Message) (proto.Message, error)

	currentGoroutineId int
	state              ConnState
	secret             []byte
	future             *Future
}

func (rc *Conn) Dial(addr string, timeout time.Duration) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	rc.address = address
	rc.timeout = timeout
	rc.state = Connecting
	rc.conn, err = net.DialTimeout("tcp", address.AddrPort().String(), timeout)
	if err != nil {
		rc.state = Disconnected
		return err
	}
	rc.state = Connected

	//1.启动发送及接收
	rc.done.Add(2)

	go rc.sender()
	go rc.reader()

	rc.mailboxDone.Add(1)
	go rc.guardian()

	return nil
}

func (rc *Conn) Redial() error {

	var err error
	rc.conn, err = net.DialTimeout("tcp", rc.address.AddrPort().String(), rc.timeout)
	if err != nil {
		return err
	}

	rc.stopper = make(chan struct{})
	rc.mailbox = make(chan interface{}, 1)

	//1.启动发送及接收
	rc.done.Add(2)

	// go rc.sender()
	go rc.reader()

	rc.mailboxDone.Add(1)
	go rc.guardian()

	return nil
}

func (rc *Conn) Register(key interface{}, f func(ctx context.Context, message proto.Message) (proto.Message, error)) {
	rc.methods[reflect.TypeOf(key)] = f
}

func (rc *Conn) IsConnected() bool {
	return rc.state == Connected
}

func (rc *Conn) ToAddress() string {
	if rc.address == nil {
		return "unknown"
	}

	return rc.address.AddrPort().String()
}

// RequestMessage 请求消息并等待回复，超时时间单位为毫秒
// proto.Message

func (rc *Conn) RequestMessage(message proto.Message, timeout int64) (client.IFuture, error) {
	if rc.currentGoroutineId == utils.GetCurrentGoroutineID() {
		panic("RequestMessage cannot block calls in its own thread")
	}

	if rc.isStopped() {
		return nil, errs.ErrorRpcConnectorClosed
	}

	// msgAny, err := anypb.New(message)
	// if err != nil {
	// 	panic(err)
	// }

	seq := rc.nextID()

	future := &Future{
		sequenceID: seq,
		cond:       sync.NewCond(&sync.Mutex{}),
		done:       false,
		request:    message,
		result:     nil,
		err:        nil,
		t:          time.NewTimer(time.Duration(timeout) * time.Millisecond),
	}

	b, err := protomessge.Marshal(message.(proto.Message), rc.secret)
	if err != nil {
		return future, nil
	}

	rc.sendcond.L.Lock()
	if rc.isStopped() {
		rc.sendcond.L.Unlock()
		return nil, errors.New("client: closed")
	}

	if _, err := rc.sendbox.WriteBinary(b); err != nil {
		rc.sendcond.L.Unlock()
		return nil, err
	}

	rc.sendcond.Signal()
	rc.sendcond.L.Unlock()

	if err != nil {
		// 发送失败
		signal := false
		future.cond.L.Lock()
		if !future.done {
			future.done = true
			future.err = err
			tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
			if tp != nil {
				tp.Stop()
			}
			// 删除映射表
			rc.waitbox.Remove(strconv.FormatInt(int64(future.sequenceID), 10))
			signal = true
		}
		future.cond.L.Unlock()
		if signal {
			future.cond.Signal()
		}

		return future, nil
	}

	rc.sendcond.L.Lock()
	node := &IntervalLinkNode{
		Value: future,
	}
	rc.waitbox.SetIfAbsent(strconv.FormatInt(int64(future.sequenceID), 10), node)

	rc.sendcond.Signal()
	rc.sendcond.L.Unlock()

	if timeout > 0 {
		tp := time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
			future.cond.L.Lock()
			if future.done {
				future.cond.L.Unlock()
				return
			}
			future.err = errs.ErrorRequestTimeout
			future.done = true

			tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
			if tp != nil {
				tp.Stop()
			}

			rc.waitbox.Remove(strconv.FormatInt(int64(future.sequenceID), 10))
			future.cond.Signal()
			future.cond.L.Unlock()
		})
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&future.t)), unsafe.Pointer(tp))
	}
	rc.future = future
	future.cond.L.Lock()
	future.cond.Wait()
	future.cond.L.Unlock()
	return future, nil
}

// func (rc *Conn) pushSendBox(msg interface{}) *IntervalLinkNode {
// 	rc.sendcon.L.Lock()
// 	// TODO: 是否已关闭
// 	node := &IntervalLinkNode{
// 		Value: msg,
// 	}
// 	rc.sendbox.Push(node)
// 	switch msgType := msg.(type) {
// 	case *Future:
// 		rc.waitbox.SetIfAbsent(strconv.FormatInt(int64(msgType.sequenceID), 10), node)
// 	default:
// 	}

// 	rc.sendcon.L.Unlock()
// 	rc.sendcon.Signal()

// 	return node
// }

// getFuture 获取并删除
func (rc *Conn) getFuture(id int32) *IntervalLinkNode {

	v, ok := rc.waitbox.Get(strconv.FormatInt(int64(id), 10))
	if !ok {
		return nil
	}

	rc.waitbox.Remove(strconv.FormatInt(int64(id), 10))
	return v.(*IntervalLinkNode)
}

func (rc *Conn) Close() {
	rc.done.Wait()
	rc.mailboxDone.Wait()
}

func (rc *Conn) Destory() {
	close(rc.mailbox)
	// rc.sendbox = nil
}

func (rc *Conn) nextID() int32 {
	for {
		if id := atomic.AddInt32(&rc.sequence, 1); id > 0 {
			return id
		} else if atomic.CompareAndSwapInt32(&rc.sequence, id, 1) {
			return 1
		}
	}
}

func (rc *Conn) isStopped() bool {
	select {
	case <-rc.stopper:
		return true
	default:
		return false
	}
}

func (rc *Conn) onResponse(msg protoreflect.ProtoMessage) {
	futureNode := rc.getFuture(rc.future.sequenceID)
	if futureNode == nil {
		return
	}

	future := futureNode.Value.(*Future)
	future.cond.L.Lock()
	if future.done {
		future.cond.L.Unlock()
		return
	}

	if msg != nil {
		rc.future.result = msg
	}

	future.done = true

	tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
	if tp != nil {
		tp.Stop()
	}
	future.cond.Signal()
	future.cond.L.Unlock()
	
}

func (c *Conn) sender() {
	defer func() {
		c.done.Done()
	}()

	var (
		err       error
		readbytes []byte = nil
	)
	for {
		c.sendcond.L.Lock()
		if !c.isStopped() {
			c.sendcond.Wait()
		}
		c.sendcond.L.Unlock()

		for {
			if c.isStopped() {
				goto tcp_sender_exit_label
			}

			c.sendcond.L.Lock()
			c.sendbox.Flush()
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

				c.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 50))
				if nwrite, err = c.conn.Write(readbytes[offset:]); err != nil {
					if e, ok := err.(net.Error); ok && e.Timeout() {
						goto tcp_sender_continue_label
					}

					goto tcp_sender_exit_label
				}
			tcp_sender_continue_label:
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

func (c *Conn) reader() {
	defer func() {
		c.done.Done()
	}()

	var tmp [512]byte
	remoteAddr := c.conn.RemoteAddr()
	for {

		if c.isStopped() {
			break
		}

		// if c.keepalive > 0 {
		// 	c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.keepalive) * time.Millisecond * 2.0))
		// }

		n, err := c.conn.Read(tmp[:])
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		// c.keepaliveError = 0
		c.mailbox <- &RecviceMessage{Data: tmp[:n], Addr: remoteAddr}
	}

	c.conn.Close()
	c.sendcond.L.Lock()
	if !c.isStopped() {
		close(c.stopper)
	}
	c.sendcond.Signal()
	c.sendcond.L.Unlock()

	c.mailbox <- &ClosedMessage{}

}

func (rc *Conn) guardian() {
	defer rc.mailboxDone.Done()

	rc.currentGoroutineId = utils.GetCurrentGoroutineID()
	if rc.Config.Connected != nil {
		rc.Config.Connected()
	}

	for {
		reqMsg, ok := <-rc.mailbox
		if !ok {
			goto exit_guardian_lable
		}

		// 直接请求退出
		if reqMsg == nil {
			break
		}
		switch msg:= reqMsg.(type) {
		case *RecviceMessage:
			rc.Recvice(msg.Data)
		case *ClosedMessage:
			goto exit_guardian_lable
		default:
		}
		
		// rc.Recvice(reqMsg.())

		// f, ok := rc.methods[reflect.TypeOf(reqMsg)]
		// if !ok {
		// 	goto exit_guardian_lable
		// }

		// result, err := f(context.Background(), reqMsg.(protoreflect.ProtoMessage))
		// if err != nil {
		// 	result = &messages.RpcError{Err: err.Error()}
		// }

	}

exit_guardian_lable:
	rc.done.Wait() // 等待读写线程结束
	rc.state = Disconnected
	rc.Config.Closed()
}
func (rc *Conn) onPingReply(message *pubs.PingMsg) {
	msg := &pubs.PingMsg{VerificationKey: message.VerificationKey + 1}
	b, err := protomessge.Marshal(msg, rc.secret)
	if err != nil {
		return 
	}

	rc.sendcond.L.Lock()
	if rc.isStopped() {
		rc.sendcond.L.Unlock()
		return 
	}

	if _, err := rc.sendbox.WriteBinary(b); err != nil {
		rc.sendcond.L.Unlock()
		return 
	}

	rc.sendcond.Signal()
	rc.sendcond.L.Unlock()
	fmt.Println("ping reply success")
}

func (dl *Conn) Recvice(b []byte) {
	offset := 0
	for {
		var (
			n    int   = 0
			werr error = nil
		)
		if offset < len(b) {
			n, werr = dl.recvice.WriteBinary(b[offset:])
			offset += n
			if err := dl.recvice.Flush(); err != nil {
				return
			}
		}

		msg, err := protomessge.UnMarshal(dl.recvice, dl.secret)
		if err != nil {
			vlog.Errorf("unmarshal message error:%v", err.Error())
			return
		}

		if msg == nil {
			if werr != nil {
				vlog.Errorf("unmarshal message error:%v", err.Error())
				return
			}

			if offset == len(b) {
				return
			}
			continue
		}

		switch message := msg.(type) {
		case *pubs.PingMsg:
			dl.onPingReply(message)
		case *pubs.PubkeyMsg:
			dl.onPubkeyReply(message)
		case *pubs.Error:
			vlog.Errorf("message %v error:%v",message.Name,message.Err)
		default:
			dl.onResponse(message)
		}
	}
}

func (dl *Conn) onPubkeyReply(message *pubs.PubkeyMsg) {

	// var (
	// 	prvKey crypto.PrivateKey
	// 	pubKey crypto.PublicKey
	// )

	// pubkeyByte, err := base64.StdEncoding.DecodeString(message.Key)
	// if err != nil {
	// 	vlog.Debugf("public key decode error %s", err.Error())
	// 	ctx.Close(ctx.Self())
	// 	return
	// }

	// prvKey, pubKey, err = dl.gateway.encryption.Ecdh.GenerateKey(rand.Reader)
	// if err != nil {
	// 	vlog.Debugf("generate public/private key error %s", err.Error())
	// }

	// remotePubkey, ok := dl.gateway.encryption.Ecdh.Unmarshal(pubkeyByte)
	// if !ok {
	// 	vlog.Debug("Public key parsing exception")
	// 	ctx.Close(ctx.Self())
	// 	return
	// }

	// secret, err := dl.gateway.encryption.Ecdh.GenerateSharedSecret(prvKey, remotePubkey)
	// if err != nil {
	// 	vlog.Debugf("generate shared secret error %s", err.Error())
	// 	ctx.Close(ctx.Self())
	// 	return
	// }

}
