package server

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/cluster/proxy/messageproxy"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/vlog"
)

func NewRpcClientConn(s *RpcServer) RpcClient {
	conn := &RpcClientConn{
		recvice:    circbuf.NewLinkBuffer(4096),
		register:   s.Register,
		unregister: s.UnRegister,
	}
	conn.processor = messages.NewRpcServiceProcessor(conn)
	conn.oprot = NewRpcContextProtocol()
	conn.message_proxy = messageproxy.NewRepeatMessageProxy()
	conn.message_proxy.(*messageproxy.RepeatMessageProxy).Register(protocol.MessageName(&messages.RpcPingMessage{}),NewRpcPingMessageProxy())
	conn.message_proxy.(*messageproxy.RepeatMessageProxy).Register(protocol.MessageName(&messages.RpcRequestMessage{}),NewRpcRequestMessageProxy(conn))
	conn.message_proxy.(*messageproxy.RepeatMessageProxy).Register(protocol.MessageName(&messages.RpcResponseMessage{}),NewRpcResponseMessageProxy(conn))
	return conn
}

type RpcClient interface {
	ClientID() *network.ClientID
	Register(key thrift.TStruct, f thrift.TProcessorFunction)
	Accept(ctx network.Context)
	Ping(ctx network.Context)
	Recvice(ctx network.Context)
	Closed(ctx network.Context)
	Destory()
	onRpcPing(ctx network.Context, iprot protocol.IProtocol)
	onRpcRequest(ctx network.Context, iprot protocol.IProtocol)
	onRpcResponse(ctx network.Context, iprot protocol.IProtocol)
	referenceIncrement() int32
	referenceDecrement() int32
}

type RpcClientConn struct {
	clientID *network.ClientID // 客户端ID
	recvice   *circbuf.LinkBuffer // 接收缓冲区
	processor thrift.TProcessor
	oprot protocol.IProtocol
	message_proxy messageproxy.IMessageProxy
	reference int32 // 引用计数器

	register   func(*network.ClientID, RpcClient)
	unregister func(*network.ClientID)
}

func (rcc *RpcClientConn) ClientID() *network.ClientID {
	return rcc.clientID
}

func (rcc *RpcClientConn) Register(key thrift.TStruct, f thrift.TProcessorFunction) {
	rcc.processor.AddToProcessorMap(protocol.MessageName(key), f)
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
		if err != nil {
			ctx.Close(ctx.Self())
			return
		}
		offset += n
		if err = rcc.recvice.Flush(); err !=nil{
			ctx.Close(ctx.Self())
			return
		}
	
		msg,err := messages.UnMarshal(rcc.recvice)
		if err !=nil{
			ctx.Close(ctx.Self())
			return
		}
		if err := rcc.message_proxy.Message(ctx,msg,0);err!= nil{
			vlog.Debugf(err.Error())
			ctx.Close(ctx.Self())
			return
		}
		// name, _, _, err := rcc.proto.ReadMessageBegin(context.Background())
		// if err != nil {
		// 	ctx.Close(ctx.Self())
		// 	return
		// }
		// switch name {
		// case "messages.RpcPingMessage":
		// 	rcc.onRpcPing(ctx, rcc.proto)
		// case "messages.RpcRequestMessage":
		// 	rcc.onRpcRequest(ctx, rcc.proto)
		// case "messages.RpcResponseMessage":
		// 	rcc.onRpcResponse(ctx, rcc.proto)
		// default:
		// 	vlog.Debug("unknown RPC message")
		// }
		// rcc.proto.Release()
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
	// rcc.recvice.Release()
}
func (rcc *RpcClientConn) onRpcRequest(ctx network.Context, iprot protocol.IProtocol) {
	request := messages.NewRpcRequestMessage()
	request.Read(context.Background(), iprot)
	iprot.Release()
	iprot.Write(request.Message)

	timeout := int64(request.Timeout) - (time.Now().UnixMilli() - int64(request.ForwardTime))
	// 如果已超时
	if timeout <= 0 {
		return
	}
	// 设置超时器
	background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	ctxx :=NewCtxWithRpcClientContext(background, &RpcClientContext{sequenceID: request.SequenceID,
		background: background,
		context:    ctx})

	rcc.processor.Process(ctxx, iprot, iprot)
	FreeCtxWithRpcClientContext(ctxx)
	// 判断是否超时
	select {
	case <-background.Done():
	default:
	}

	cancel()
}
func (rcc *RpcClientConn) onRpcResponse(ctx network.Context, iprot protocol.IProtocol) {
	response := messages.NewRpcResponseMessage()
	response.Read(context.Background(), iprot)
	iprot.Release()
	iprot.Write(response.Result_)
	background := context.Background()
	ctxx :=NewCtxWithRpcClientContext(background, &RpcClientContext{sequenceID: response.SequenceID,
		background: background,
		context:    ctx})
	rcc.processor.Process(ctxx, iprot, iprot)
	FreeCtxWithRpcClientContext(ctxx)
}

func (rcc *RpcClientConn) onRpcPing(ctx network.Context, iprot protocol.IProtocol) {
	ping := messages.NewRpcPingMessage()
	ping.Read(context.Background(), iprot)
	iprot.Release()
	iprot.WriteMessageBegin(context.Background(), protocol.MessageName(ping), thrift.EXCEPTION, 1)
	ping.VerifyKey = ping.VerifyKey + 1
	ping.Write(context.Background(), iprot)
	ctx.PostMessage(ctx.Self(), iprot.GetBytes())
}

// RefInc 引用计数器+1
func (rcc *RpcClientConn) referenceIncrement() int32 {
	return atomic.AddInt32(&rcc.reference, 1)
}

// RefDec 引用计数器-1
func (rcc *RpcClientConn) referenceDecrement() int32 {
	return atomic.AddInt32(&rcc.reference, -1)
}
