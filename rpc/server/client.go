package server

import (
	"context"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	rpcmessage "github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewRpcClientConn(s *RpcServer) RpcClient {
	return &RpcClientConn{
		recvice:    circbuf.NewLinkBuffer(4096),
		methods:    make(map[interface{}]func(*RpcClientContext) (protoreflect.ProtoMessage, error)),
		register:   s.Register,
		unregister: s.UnRegister,
	}
}

type RpcClient interface {
	ClientID() *network.ClientID
	Register(key interface{}, f func(*RpcClientContext) (protoreflect.ProtoMessage, error))
	Accept(ctx network.Context)
	Ping(ctx network.Context)
	Recvice(ctx network.Context)
	Closed(ctx network.Context)
	Destory()
	onRpcPing(ctx network.Context, message *rpcmessage.RpcPingMessage)
	referenceIncrement() int32
	referenceDecrement() int32
}

type RpcClientConn struct {
	clientID  *network.ClientID   // 客户端ID
	recvice   *circbuf.LinkBuffer // 接收缓冲区
	methods   map[interface{}]func(*RpcClientContext) (proto.Message, error)
	reference int32 // 引用计数器

	register   func(*network.ClientID, RpcClient)
	unregister func(*network.ClientID)
}

func (rcc *RpcClientConn) ClientID() *network.ClientID {
	return rcc.clientID
}

func (rcc *RpcClientConn) Register(key interface{}, f func(*RpcClientContext) (protoreflect.ProtoMessage, error)) {
	rcc.methods[reflect.TypeOf(key)] = f
}

func (rcc *RpcClientConn) Accept(ctx network.Context) {
	rcc.clientID = ctx.Self()
	rcc.reference = 1
	rcc.register(ctx.Self(), rcc)
}

func (rcc *RpcClientConn) Ping(ctx network.Context) {
	// 不主动处理心跳
}

func (rcc *RpcClientConn) Recvice(ctx network.Context) {
	// 通用化需要修改
	offset := 0
	for {
		n, err := rcc.recvice.WriteBinary(ctx.Message()[offset:])
		offset += n

		_, msg, msgErr := messages.UnMarshalProtobuf(rcc.recvice)
		if msgErr != nil {
			ctx.Close(ctx.Self())
			return
		}

		if msg == nil && err != nil {
			// 属于缓冲区溢出
			ctx.Close(ctx.Self())
			return
		}

		if msg == nil {
			goto rpc_client_offset_label
		}

		switch message := msg.(type) {
		case *messages.RpcPingMessage:
			rcc.onRpcPing(ctx, message)
		case *messages.RpcRequestMessage:
			reqMsg, err := message.Message.UnmarshalNew()
			if err != nil {
				goto rpc_client_offset_label
			}

			f, ok := rcc.methods[reflect.TypeOf(reqMsg)]
			if !ok {
				goto rpc_client_offset_label
			}

			timeout := int64(message.Timeout) - (time.Now().UnixMilli() - int64(message.ForwardTime))
			// 如果已超时
			if timeout <= 0 {
				goto rpc_client_offset_label
			}
			// 设置超时器
			background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
			result, err := f(&RpcClientContext{sequenceID: message.SequenceID,
				background: background,
				context:    ctx,
				message:    reqMsg,
			})

			// 判断是否超时
			select {
			case <-background.Done():
			default:
				var b []byte
				if err != nil {
					b, _ = messages.MarshalResponseProtobuf(message.SequenceID, &rpcmessage.RpcError{Err: err.Error()})
				} else {
					b, _ = messages.MarshalResponseProtobuf(message.SequenceID, result)
				}
				if b != nil {
					ctx.PostMessage(ctx.Self(), b)
				}
			}

			cancel()
		case *messages.RpcResponseMessage:
			/*postMsg, err := message.Message.UnmarshalNew()
			if err != nil {
				goto rpc_client_offset_label
			}
			f, ok := rcc.methods[reflect.TypeOf(postMsg)]
			if !ok {
				goto rpc_client_offset_label
			}

			f(&RpcClientContext{context: ctx,
				message: postMsg})*/

		default:
			vlog.Debug("unknown RPC message")
		}
	rpc_client_offset_label:
		if offset == len(ctx.Message()) {
			break
		}
	}
}

func (rcc *RpcClientConn) Closed(ctx network.Context) {
	rcc.unregister(ctx.Self())
}

func (rcc *RpcClientConn) Destory() {
	rcc.clientID = nil
	rcc.recvice.Release()
}

func (rcc *RpcClientConn) onRpcPing(ctx network.Context, message *messages.RpcPingMessage) {

	b, err := messages.MarshalPingProtobuf(message.VerifyKey + 1)
	if err != nil {
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), b)
}

// RefInc 引用计数器+1
func (rcc *RpcClientConn) referenceIncrement() int32 {
	return atomic.AddInt32(&rcc.reference, 1)
}

// RefDec 引用计数器-1
func (rcc *RpcClientConn) referenceDecrement() int32 {
	return atomic.AddInt32(&rcc.reference, -1)
}
