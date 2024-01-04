package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewConn(options ...ConnConfigOption) *Conn {
	config := Configure(options...)

	return NewConnConfig(config)
}

func NewConnConfig(config *ConnConfig) *Conn {
	return &Conn{
		Config:   config,
		sendbox:  intrusive.NewLinked(&syncx.NoMutex{}),
		sendcon:  sync.NewCond(&sync.Mutex{}),
		waitbox:  cmap.New(),
		methods:  make(map[interface{}]func(ctx context.Context, message proto.Message) (proto.Message, error)),
		stopper:  make(chan struct{}),
		sequence: 1,
		mailbox:  make(chan interface{}, 1),
		state:    Disconnected,
	}
}

type ConnState int

const (
	Disconnected = iota
	Connecting
	Connected
	Disconnecting
)

type IntervalLinkNode struct{
	intrusive.LinkedNode
	Value interface{}
}

type Conn struct {
	BaseConnect
	Config  *ConnConfig
	conn    net.Conn
	address *net.TCPAddr
	timeout time.Duration
	sendbox *intrusive.Linked
	waitbox cmap.ConcurrentMap
	sendcon *sync.Cond

	stopper  chan struct{}
	sequence int32
	done     sync.WaitGroup

	mailbox     chan interface{}
	mailboxDone sync.WaitGroup

	methods map[interface{}]func(ctx context.Context, message proto.Message) (proto.Message, error)

	ping               uint64
	kleepaliveError    int32
	currentGoroutineId int
	state              ConnState
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

	go rc.sender()
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
func (rc *Conn) RequestMessage(message proto.Message, timeout int64) (*Future, error) {
	if rc.currentGoroutineId == utils.GetCurrentGoroutineID() {
		panic("RequestMessage cannot block calls in its own thread")
	}

	if rc.isStopped() {
		return nil, errs.ErrorRpcConnectorClosed
	}

	msgAny, err := anypb.New(message)
	if err != nil {
		panic(err)
	}

	seq := rc.nextID()
	req := &messages.RpcRequestMessage{
		SequenceID:  seq,
		ForwardTime: uint64(time.Now().UnixMilli()),
		Timeout:     uint64(timeout),
		Message:     msgAny,
	}

	future := &Future{
		sequenceID: seq,
		cond:       sync.NewCond(&sync.Mutex{}),
		done:       false,
		request:    req,
		result:     nil,
		err:        nil,
		t:          time.NewTimer(time.Duration(timeout) * time.Millisecond),
	}

	futureNode := rc.pushSendBox(future)
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

			rc.sendbox.Remove(futureNode)
			rc.waitbox.Remove(strconv.FormatInt(int64(future.sequenceID), 10))

			future.cond.L.Unlock()
			future.cond.Signal()
		})
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&future.t)), unsafe.Pointer(tp))
	}

	return future, nil
}

func (rc *Conn) responseMessage(sequenceID int32, message proto.Message) error {
	var (
		resultAny *anypb.Any
		err       error
	)

	if rc.isStopped() {
		return errs.ErrorRpcConnectorClosed
	}

	if message != nil {
		resultAny, err = anypb.New(message)
		if err != nil {
			panic(err)
		}
	}

	rc.pushSendBox(&messages.RpcResponseMessage{
		SequenceID: sequenceID,
		Result:     resultAny,
	})

	return nil
}

func (rc *Conn) pushSendBox(msg interface{}) *IntervalLinkNode {
	rc.sendcon.L.Lock()
	// TODO: 是否已关闭
	node := &IntervalLinkNode{
		Value: msg,
	}
	rc.sendbox.Push(node)
	switch msgType := msg.(type) {
	case *Future:
		rc.waitbox.SetIfAbsent(strconv.FormatInt(int64(msgType.sequenceID), 10), node)
	default:
	}

	rc.sendcon.L.Unlock()
	rc.sendcon.Signal()

	return node
}

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
	rc.sendcon.L.Lock()
	rc.sendbox.Push(nil)
	rc.sendcon.L.Unlock()
	rc.sendcon.Signal()
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

func (rc *Conn) onResponse(msg *messages.RpcResponseMessage) {
	futureNode := rc.getFuture(msg.SequenceID)
	if futureNode == nil {
		return
	}

	future := futureNode.Value.(*Future)
	future.cond.L.Lock()
	if future.done {
		future.cond.L.Unlock()
		return
	}

	if msg.Result != nil {
		result, err := msg.Result.UnmarshalNew()
		if err != nil {
			future.cond.L.Unlock()
			panic(err)
		}

		future.result = result
		switch resultType := result.(type) {
		case *messages.RpcError:
			future.err = errors.New(resultType.Err)
		default:
			future.err = nil
		}
	}

	future.done = true

	tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
	if tp != nil {
		tp.Stop()
	}

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

			var (
				b   []byte
				err error
			)

			rc.sendcon.L.Lock()
			msgNode := rc.sendbox.Pop()
			rc.sendcon.L.Unlock()

			if msgNode.(*IntervalLinkNode).Value == nil {
				goto exit_sender_lable
			}

			switch msg := msgNode.(*IntervalLinkNode).Value.(type) {
			case *Future:
				msg.cond.L.Lock()
				if msg.done {
					msg.cond.L.Unlock()
					continue
				}

				// 剩余时间小于超时20%无再发送意义,直接等待超时
				diff := int64(msg.request.Timeout) - (time.Now().UnixMilli() - int64(msg.request.ForwardTime))
				if diff < int64(float64(msg.request.Timeout)*0.2) {
					msg.cond.L.Unlock()
					continue
				}
				msg.cond.L.Unlock()

				b, err = messages.MarshalRequestProtobuf(msg.request.SequenceID, msg.request.Timeout, msg.request.Message)
				if err != nil {
					goto exit_sender_lable
				}

				_, err = rc.conn.Write(b)
				if err != nil {
					// 发送失败
					signal := false
					msg.cond.L.Lock()
					if !msg.done {
						msg.done = true
						msg.err = err
						tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&msg.t))))
						if tp != nil {
							tp.Stop()
						}
						// 删除映射表
						rc.waitbox.Remove(strconv.FormatInt(int64(msg.sequenceID), 10))
						signal = true
					}
					msg.cond.L.Unlock()
					if signal {
						msg.cond.Signal()
					}

					goto exit_sender_lable
				}
			case *messages.RpcResponseMessage:
				b, err = messages.MarshalResponseProtobuf(msg.SequenceID, msg.Result)
				if err != nil {
					goto exit_sender_lable
				}

				_, err = rc.conn.Write(b)
				if err != nil {
					goto exit_sender_lable
				}
			case *messages.RpcPingMessage:
				b, err = messages.MarshalPingProtobuf(msg.VerifyKey)
				if err != nil {
					goto exit_sender_lable
				}
				_, err = rc.conn.Write(b)
				if err != nil {
					goto exit_sender_lable
				}
			default:
				panic("sender: unknown rpc message")
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
	readbuffer := circbuf.NewLinkBuffer(4096)
	defer readbuffer.Close()

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
				pingMessage := &messages.RpcPingMessage{VerifyKey: rc.ping}
				rc.pushSendBox(pingMessage)
				continue
			}
			goto exit_reader_lable
		}

		rc.kleepaliveError = 0
		offset := 0
		for {
			nw, err := readbuffer.WriteBinary(readtemp[offset:nr])
			offset += nw

			_, msg, uerr := messages.UnMarshalProtobuf(readbuffer)
			if uerr != nil {
				goto exit_reader_lable
			}

			if err != nil && msg == nil {
				// 数据包存在问题
				goto exit_reader_lable
			}

			if msg != nil {
				switch message := msg.(type) {
				case *messages.RpcPingMessage:

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

	rc.currentGoroutineId = utils.GetCurrentGoroutineID()
	if rc.Config.Connected != nil {
		rc.Config.Connected()
	}

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
		case *messages.RpcResponseMessage:
			rc.onResponse(message)
		case *messages.RpcRequestMessage:

			reqMsg, err := message.Message.UnmarshalNew()
			if err != nil {
				panic(err)
			}

			f, ok := rc.methods[reflect.TypeOf(reqMsg)]
			if !ok {
				goto exit_guardian_lable
			}

			timeout := int64(message.Timeout) - (time.Now().UnixMilli() - int64(message.ForwardTime))
			if timeout <= 0 {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
			result, err := f(ctx, reqMsg)
			if err != nil {
				result = &messages.RpcError{Err: err.Error()}
			}

			select {
			case <-ctx.Done(): // 已超时
			default:
				rc.responseMessage(message.SequenceID, result)
			}

			cancel()
		default:
			panic(fmt.Errorf("unknown %s message", reflect.TypeOf(message).Name()))
		}
	}

exit_guardian_lable:
	rc.done.Wait() // 等待读写线程结束
	rc.state = Disconnected
	rc.BaseConnect.Affiliation().Remove(rc.node)

	for {
		rc.sendcon.L.Lock()
		sendMsgNode := rc.sendbox.Pop()
		rc.sendcon.L.Unlock()

		if sendMsgNode == nil {
			break
		}

		switch sendMsg := sendMsgNode.(*IntervalLinkNode).Value.(type) {
		case *Future:
			sendMsg.cond.L.Lock()
			if !sendMsg.done {
				sendMsg.done = true
				sendMsg.err = errs.ErrorRpcConnectorClosed
				tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&sendMsg.t))))
				if tp != nil {
					tp.Stop()
				}
			}
			sendMsg.cond.L.Unlock()
			sendMsg.cond.Signal()
		default:
		}

		sendMsgNode.(*IntervalLinkNode).Value = nil
	}

	/*rc.requestbox.Range(func(key, value any) bool {
		future := value.(*Future)
		future.cond.L.Lock()
		if future.done {
			future.cond.L.Unlock()
			return true
		}

		future.done = true
		future.result = nil
		future.err = errs.ErrorRpcConnectorClosed
		future.done = true
		future.cond.L.Unlock()
		future.cond.Signal()

		rc.removeFuture(key.(int32))

		return true
	})*/
	rc.Config.Closed()
}
