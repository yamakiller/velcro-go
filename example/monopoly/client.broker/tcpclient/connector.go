package tcpclient

import (
	"context"
	"net"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/rpc/client"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewConn(options ...client.ConnConfigOption) client.IConnect {
	config := client.Configure(options...)

	return NewConnConfig(config)
}

func NewConnConfig(config *client.ConnConfig) *Conn {
	return &Conn{
		BaseConnect: &client.BaseConnect{},
		Config:      config,
		// sendbox:     intrusive.NewLinked(&syncx.NoMutex{}),
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

type IntervalLinkNode struct {
	intrusive.LinkedNode
	Value interface{}
}

type Conn struct {
	*client.BaseConnect
	Config  *client.ConnConfig
	conn    net.Conn
	address *net.TCPAddr
	timeout time.Duration
	// sendbox *intrusive.Linked
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

	// go rc.sender()
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

	rc.sendcon.L.Lock()
	_, err = rc.conn.Write(b)
	rc.sendcon.L.Unlock()
	rc.sendcon.Signal()

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

	rc.sendcon.L.Lock()
	node := &IntervalLinkNode{
		Value: future,
	}
	rc.waitbox.SetIfAbsent(strconv.FormatInt(int64(future.sequenceID), 10), node)
	rc.sendcon.L.Unlock()
	rc.sendcon.Signal()

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

			// rc.sendbox.Remove(futureNode)
			rc.waitbox.Remove(strconv.FormatInt(int64(future.sequenceID), 10))

			future.cond.L.Unlock()
			future.cond.Signal()
		})
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&future.t)), unsafe.Pointer(tp))
	}
	rc.future = future
	future.cond.L.Lock()
	future.cond.Wait()
	future.cond.L.Unlock()
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

	req := &messages.RpcResponseMessage{
		SequenceID: sequenceID,
		Result:     resultAny,
	}

	b, err := messages.MarshalResponseProtobuf(req.SequenceID, req.Result)
	if err != nil {
		return errs.ErrorRpcConnectorClosed
	}
	rc.sendcon.L.Lock()
	_, err = rc.conn.Write(b)
	rc.sendcon.L.Unlock()
	rc.sendcon.Signal()
	if err != nil {
		return errs.ErrorRpcConnectorClosed
	}

	return nil
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
	// rc.sendcon.L.Lock()
	// rc.sendbox.Push(&IntervalLinkNode{Value: nil})
	// rc.sendcon.L.Unlock()
	// rc.sendcon.Signal()
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
	future.cond.L.Unlock()
	future.cond.Signal()
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
				// pingMessage := &messages.RpcPingMessage{VerifyKey: rc.ping}
				b, err := messages.MarshalPingProtobuf(rc.ping)
				if err != nil {
					goto exit_reader_lable
				}
				_, err = rc.conn.Write(b)
				if err != nil {
					goto exit_reader_lable
				}
				continue
			}
			goto exit_reader_lable
		}

		rc.kleepaliveError = 0
		offset := 0
		for {
			nw, err := readbuffer.WriteBinary(readtemp[offset:nr])
			if err != nil {
				goto exit_reader_lable
			}
			offset += nw
			if err := readbuffer.Flush(); err != nil {
				goto exit_reader_lable
			}
			msg, uerr := protomessge.UnMarshal(readbuffer, rc.secret)
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
		reqMsg, ok := <-rc.mailbox
		if !ok {
			goto exit_guardian_lable
		}

		// 直接请求退出
		if reqMsg == nil {
			break
		}

		rc.onResponse(reqMsg.(protoreflect.ProtoMessage))

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
	rc.BaseConnect.Affiliation().Remove(rc.Node())

	rc.Config.Closed()
}
