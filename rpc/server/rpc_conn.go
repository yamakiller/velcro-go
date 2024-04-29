package server

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/stringx"
	"github.com/yamakiller/velcro-go/vlog"
)

/*func NewRpcClientConn(s *RpcServer) RpcClient {
	conn := &RpcClientConn{
		recvice:    circbuf.NewLinkBuffer(32),
		register:   s.Register,
		unregister: s.UnRegister,
	}
	conn.processor = messages.NewRpcServiceProcessor(conn)
	conn.oprot = NewRpcContextProtocol()
	conn.message_agent = NewRpcClientMessageAgent(conn)

	return conn
}*/

type IRpcConnector interface {
	GetID() *network.ClientID
	Destory() error

	Accept(network.Context)
	Ping(network.Context)
	Recvice(network.Context)
	Closed(network.Context)

	referenceIncrement() int32
	referenceDecrement() int32
}

type RpcConnector struct {
	id      *network.ClientID
	recvice *circbuf.LinkBuffer // 接收缓冲区
	//thrift        thrift.TProcessor   // thrift 处理器
	thriftHandler protocol.IProtocol
	managerAgent  IRpcManagerConnector

	reference int32 // 引用计数器
}

func (rc *RpcConnector) GetID() *network.ClientID {
	return rc.id
}

func (rc *RpcConnector) Destory() error {
	var err error
	if rc.id != nil {
		err = rc.id.UserClose()
		rc.id = nil
	}

	return err
}

func (rc *RpcConnector) Accept(ctx network.Context) {
	rc.id = ctx.Self()
	rc.reference = 1

	rc.managerAgent.Register(ctx.Self(), rc)
}

func (rc *RpcConnector) Ping(ctx network.Context) {

}

func (rc *RpcConnector) Recvice(ctx network.Context) {
	offset := 0
	for {
		n, err := rc.recvice.WriteBinary(ctx.Message()[offset:])
		if err != nil {
			ctx.Close(ctx.Self())
			return
		}
		offset += n
		if err = rc.recvice.Flush(); err != nil {
			ctx.Close(ctx.Self())
			return
		}

		msg, err := messages.UnMarshal(rc.recvice)
		if err != nil {
			ctx.Close(ctx.Self())
			return
		}

		if err := rc.protocolHandler(ctx, msg, 0); err != nil {
			vlog.Debugf(err.Error())
			ctx.Close(ctx.Self())
			return
		}

		if offset == len(ctx.Message()) {
			break
		}
	}
}

func (rc *RpcConnector) Closed(ctx network.Context) {
	rc.managerAgent.Unregister(ctx.Self())
}

func (rc *RpcConnector) protocolHandler(ctx network.Context, msg []byte, timeout int64) error {
	rc.thriftHandler.Release()
	rc.thriftHandler.Write(msg)

	name, _, _, err := rc.thriftHandler.ReadMessageBegin(context.Background())
	if err != nil {
		return err
	}

	hashVal := stringx.StrToHash(name)
	//TODO: 加入超时计算

	switch hashVal {
	case requestMessageID:

		//var outMsg thrift.TStruct
		req := &messages.RpcRequestMessage{}
		if err := req.Read(context.Background(), rc.thriftHandler); err != nil {
			return err
		}

		if err := rc.thriftHandler.ReadMessageEnd(context.Background()); err != nil {
			return err
		}

		timeout := int64(req.Timeout) - (time.Now().UnixMilli() - int64(req.ForwardTime))
		if timeout <= 0 {
			// TODO: 超时什么都不做
			return nil
		}

		rc.thriftHandler.Release()
		rc.thriftHandler.Write(req.Message)

		// 解码 req.Message
		//reqName, _, _, err := rc.thriftHandler.ReadMessageBegin(context.Background())
		if err != nil {
			return err
		}
		//

		// 执行请求
		//rc.thrift.Process(context.Background(), outMsg)
		break
	case pingMessageID:
		break
	default:
		return errs.ErrorRpcUnknownMessage
	}

	return nil
}

// RefInc 引用计数器+1
func (rc *RpcConnector) referenceIncrement() int32 {
	return atomic.AddInt32(&rc.reference, 1)
}

// RefDec 引用计数器-1
func (rc *RpcConnector) referenceDecrement() int32 {
	return atomic.AddInt32(&rc.reference, -1)
}

/*type RpcClient interface {
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
	clientID  *network.ClientID   // 客户端ID
	recvice   *circbuf.LinkBuffer // 接收缓冲区
	processor thrift.TProcessor
	oprot     protocol.IProtocol
	//message_agent message.IMessageAgent
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
		if err = rcc.recvice.Flush(); err != nil {
			ctx.Close(ctx.Self())
			return
		}

		msg, err := messages.UnMarshal(rcc.recvice)
		if err != nil {
			ctx.Close(ctx.Self())
			return
		}
		if err := rcc.message_agent.Message(ctx, msg, 0); err != nil {
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
	ctxx := NewCtxWithRpcClientContext(background, &RpcClientContext{sequenceID: request.SequenceID,
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
	ctxx := NewCtxWithRpcClientContext(background, &RpcClientContext{sequenceID: response.SequenceID,
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
*/
