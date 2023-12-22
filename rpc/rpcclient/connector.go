package rpcclient

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/yamakiller/velcro-go/containers"
	"github.com/yamakiller/velcro-go/rpc"
	"github.com/yamakiller/velcro-go/rpc/rpcmessage"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
)

func NewConn(options ...ConnConfigOption) *Conn {
	config := Configure(options...)

	return NewConnConfig(config)
}

func NewConnConfig(config *ConnConfig) *Conn {
	return &Conn{
		Config:     config,
		sendbox:    containers.NewQueue(8, &syncx.NoMutex{}),
		sendcon:    sync.NewCond(&sync.Mutex{}),
		stopper:    make(chan struct{}),
		sequence:   1,
		mailbox:    make(chan interface{}, 1),
		requestbox: sync.Map{},
		state:      Disconnected,
	}
}

type ConnState int

const (
	Disconnected = iota
	Connecting
	Connected
	Disconnecting
)

type ConnectedFunc func()
type ReceiveFunc func(msg interface{})
type ClosedFunc func()

type Conn struct {
	Config     *ConnConfig
	conn       net.Conn
	address    *net.TCPAddr
	timeout    time.Duration
	sendbox    *containers.Queue
	sendcon    *sync.Cond
	requestbox sync.Map
	stopper    chan struct{}
	sequence   int32
	done       sync.WaitGroup

	mailbox     chan interface{}
	mailboxDone sync.WaitGroup

	ping            uint64
	kleepaliveError int32

	state ConnState
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

	rc.Config.Connected()

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

	go rc.sender()
	go rc.reader()

	rc.mailboxDone.Add(1)
	go rc.guardian()

	rc.Config.Connected()

	return nil
}

func (rc *Conn) IsConnected() bool {
	if rc.state == Connected {
		return true
	}

	return false
}

func (rc *Conn) ToAddress() string {
	if rc.address == nil {
		return "unknown"
	}

	return rc.address.AddrPort().String()
}

// RequestMessage 请求消息并等待回复，超时时间单位为毫秒
func (rc *Conn) RequestMessage(message interface{}, timeout uint64) (interface{}, error) {
	seq := rc.nextID()
	req := &rpcmessage.RpcRequestMessage{
		SequenceID:  seq,
		ForwardTime: uint64(time.Now().UnixMilli()),
		Timeout:     timeout,
		Message:     message,
	}

	future := &Future{
		sequenceID: seq,
		cond:       sync.NewCond(&sync.Mutex{}),
		done:       false,
		result:     nil,
		err:        nil,
		t:          time.NewTimer(time.Duration(timeout) * time.Millisecond),
	}

	if err := rc.pushSendBox(req, future); err != nil {
		return nil, err
	}

	if timeout > 0 {
		tp := time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
			future.cond.L.Lock()
			if future.done {
				future.cond.L.Unlock()

				return
			}
			future.err = rpc.ErrorRequestTimeout
			future.cond.L.Unlock()
			future.Stop(rc)
		})
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&future.t)), unsafe.Pointer(tp))
	}

	future.wait()

	return future.result, future.err
}

// 推送消息
func (rc *Conn) PostMessage(message interface{}) error {
	return rc.pushSendBox(&rpcmessage.RpcMsgMessage{
		SequenceID: rc.nextID(),
		Message:    message,
	}, nil)
}

func (rc *Conn) pushSendBox(data interface{}, future *Future) error {
	rc.sendcon.L.Lock()
	if rc.isStopped() {
		rc.sendcon.L.Unlock()
		return errors.New("rpc connector: closed")
	}

	rc.sendbox.Push(data)
	if future != nil {
		rc.requestbox.Store(future.sequenceID, future)
	}

	rc.sendcon.L.Unlock()
	rc.sendcon.Signal()

	return nil
}

func (rc *Conn) getFuture(id int32) *Future {
	v, ok := rc.requestbox.Load(id)
	if !ok {
		return nil
	}

	return v.(*Future)
}

func (rc *Conn) removeFuture(id int32) {
	rc.requestbox.Delete(id)
}

func (rc *Conn) Close() {
	rc.conn.Close()
	rc.done.Wait()
	rc.mailboxDone.Wait()
}

func (rc *Conn) Destory() {
	close(rc.mailbox)
	rc.sendbox = nil
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

func (rc *Conn) onResponse(msg *rpcmessage.RpcResponseMessage) {
	future := rc.getFuture(msg.SequenceID)
	if future == nil {
		return
	}

	future.cond.L.Lock()
	if future.done {
		future.cond.L.Unlock()
		return
	}

	future.result = msg.Message
	if msg.Result == -1 {
		future.err = rpc.ErrorRequestTimeout
	} else {
		future.err = nil
	}

	future.done = true

	tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
	if tp != nil {
		tp.Stop()
	}

	rc.removeFuture(msg.SequenceID)
	future.cond.L.Unlock()

	future.cond.Signal()
}

func (rc *Conn) sender() {
	defer rc.done.Done()

	for {
		rc.sendcon.L.Lock()
		if !rc.isStopped() {
			rc.sendcon.Wait()
		}
		rc.sendcon.L.Unlock()

		for {
			if rc.isStopped() {
				goto exit_sender_lable
			}

			msg, ok := rc.sendbox.Pop()
			if ok && msg == nil {
				goto exit_sender_lable
			}

			if !ok {
				break
			}

			var b []byte
			var err error
			switch message := msg.(type) {
			case *rpcmessage.RpcRequestMessage:
				// TODO: 检测是否超时
				future := rc.getFuture(message.SequenceID)
				// 这个请求已失败,不需要再执行
				if future == nil {
					continue
				}

				future.cond.L.Lock()
				// 请求已完成
				if future.done {
					future.cond.L.Unlock()
					continue
				}

				// 剩余时间小于超时20%无再发送意义,直接等待超时
				diff := int64(message.Timeout) - (time.Now().UnixMilli() - int64(message.ForwardTime))
				if diff < int64(float64(message.Timeout)*0.2) {
					future.cond.L.Unlock()
					continue
				}
				future.cond.L.Unlock()

				b, err = rc.Config.MarshalRequest(message.SequenceID, message.Timeout, message.Message)
			case *rpcmessage.RpcMsgMessage:
				b, err = rc.Config.MarshalMessage(message.SequenceID, message.Message)
			case *rpcmessage.RpcPingMessage:
				b, err = rc.Config.MarshalPing(message.VerifyKey)
				if err != nil {
					goto exit_sender_lable
				}

			default:
				panic("sender: unknown rpc message")
			}

			if err != nil {
				goto exit_sender_lable
			}

			_, err = rc.conn.Write(b)
			if err != nil {
				goto exit_sender_lable
			}
		}
	}
exit_sender_lable:
	rc.state = Disconnecting
	rc.conn.Close()
}

func (rc *Conn) reader() {
	defer rc.done.Done()

	var readtemp [1024]byte
	readbuffer := circbuf.New(32768, &syncx.NoMutex{})
	for {
		if rc.isStopped() {
			goto exit_reader_lable
		}

		if rc.Config.Kleepalive > 0 {
			rc.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(rc.Config.Kleepalive)))
		}

		nr, err := rc.conn.Read(readtemp[:])
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				rc.kleepaliveError++
				if rc.kleepaliveError > 3 {
					goto exit_reader_lable
				}

				//发送心跳
				uid := uuid.New()
				rc.ping, _ = strconv.ParseUint(uid.String(), 10, 64)
				pingMessage := &rpcmessage.RpcPingMessage{VerifyKey: rc.ping}
				rc.pushSendBox(pingMessage, nil)
				continue
			}
			goto exit_reader_lable
		}

		rc.kleepaliveError = 0
		offset := 0
		for {
			nw, err := readbuffer.Write(readtemp[offset:nr])
			offset += nw

			_, msg, uerr := rc.Config.UnMarshal(readbuffer)
			if uerr != nil {
				goto exit_reader_lable
			}

			if err != nil && msg == nil {
				// 数据包存在问题
				goto exit_reader_lable
			}

			if msg != nil {
				switch message := msg.(type) {
				case *rpcmessage.RpcPingMessage:
				default:
					rc.mailbox <- message
				}
			}

			if offset == nr {
				break
			}
		}

	}
exit_reader_lable:
	rc.conn.Close()
	close(rc.stopper)
	rc.sendcon.Signal()
	rc.state = Disconnecting
	rc.mailbox <- nil
}

func (rc *Conn) guardian() {
	defer rc.mailboxDone.Done()

	for {
		msg, ok := <-rc.mailbox
		if !ok {
			goto exit_guardian_lable
		}

		// 直接请求退出
		if msg == nil {
			break
		}

		switch message := msg.(type) {
		case *rpcmessage.RpcResponseMessage:
			rc.onResponse(message)
		case *rpcmessage.RpcMsgMessage:
			rc.Config.Receive(message.Message)
		}
	}

exit_guardian_lable:
	rc.done.Wait() // 等待读写线程结束
	rc.state = Disconnected
	rc.requestbox.Range(func(key, value any) bool {
		future := value.(*Future)
		future.cond.L.Lock()
		if future.done {
			future.cond.L.Unlock()
			return true
		}

		future.done = true
		future.result = nil
		future.err = rpc.ErrorRpcClientClosed
		future.done = true
		future.cond.L.Unlock()
		future.cond.Signal()

		rc.removeFuture(key.(int32))

		return true
	})
	rc.Config.Closed()
}
