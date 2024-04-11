package serve

import (
	"context"
	"math"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
	"github.com/yamakiller/velcro-go/vlog"
	// "google.golang.org/protobuf/proto"
)

// Background context.Context

// ServantClientContext

type ServantClientConn struct {
	Servant *Servant
	actor   ServantClientActor
	recvice  *circbuf.LinkBuffer
	iprot     protocol.IProtocol
	oprot     protocol.IProtocol
	processor thrift.TProcessor
	// events  map[interface{}]interface{}
}

func (c *ServantClientConn) Accept(ctx network.Context) {
	c.incarnateActor()

	c.Servant.clientMutex.Lock()
	defer c.Servant.clientMutex.Unlock()
	c.Servant.clients[network.Key(ctx.Self())] = ctx.Self()
}

// Recvice 接收到的数据
func (c *ServantClientConn) Recvice(ctx network.Context) {
	offset := 0
	for {
		if offset < len(ctx.Message()) {
			n, werr := c.recvice.WriteBinary(ctx.Message()[offset:])
			if werr != nil{
				ctx.Close(ctx.Self())
				return
			}
			offset += n
			if err := c.recvice.Flush(); err != nil {
				ctx.Close(ctx.Self())
				return
			}
		}

		msg, err := messages.UnMarshal(c.recvice)
		if err != nil {
			vlog.Errorf("unmarshal message error:%v", err.Error())
			ctx.Close(ctx.Self())
			return
		}
		c.iprot.Release()
		c.iprot.Write(msg)
		name, _, _, _ := c.iprot.ReadMessageBegin(context.Background())
		switch name {
		case "messages.RpcPingMessage":
			c.onRpcPing(ctx, c.iprot)
		case "messages.RpcRequestMessage":
			c.onRpcRequest(ctx, c.iprot)
		default:
			vlog.Errorf("Unknown service:%v", name)
			ctx.Close(ctx.Self())
			return
		}

		// _, msg, msgErr := messages.UnMarshalProtobuf(c.recvice)
		// if msgErr != nil {
		// 	ctx.Close(ctx.Self())
		// 	return
		// }

		// if msg == nil && err != nil {
		// 	// 属于缓冲区溢出
		// 	ctx.Close(ctx.Self())
		// 	return
		// }

		// if msg == nil {
		// 	goto servant_client_offset_label
		// }

		// switch message := msg.(type) {
		// case *messages.RpcPingMessage:
		// 	c.onRpcPing(ctx, message)
		// case *messages.RpcRequestMessage:
		// 	reqMsg, err := message.Message.UnmarshalNew()
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	//TODO: 这里需要抛给并行器
		// 	timeout := int64(message.Timeout) - (time.Now().UnixMilli() - int64(message.ForwardTime))
		// 	// 如果已超时
		// 	if timeout <= 0 {
		// 		goto servant_client_offset_label
		// 	}
		// 	msgType, err := protoregistry.GlobalTypes.FindMessageByName(reqMsg.(*anypb.Any).MessageName())
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	m := msgType.New().Interface()
		// 	proto.Unmarshal(reqMsg.(*anypb.Any).GetValue(), m)
		// 	reqMsg = m

		// 	var messageEnvelope *MessageEnvelope
		// 	switch rs := reqMsg.(type) {
		// 	case *prvs.ForwardBundle:
		// 		reqMsg, err = rs.Body.UnmarshalNew()
		// 		if err != nil {
		// 			panic(err)
		// 		}
		// 		messageEnvelope = NewMessageEnvelopePool(message.SequenceID, rs.Sender, reqMsg)
		// 	default:
		// 		messageEnvelope = NewMessageEnvelopePool(message.SequenceID, nil, reqMsg)
		// 	}

		// 	evt, ok := c.events[reflect.TypeOf(reqMsg)]
		// 	if !ok || evt.(func(context.Context) (proto.Message, error)) == nil {
		// 		panic(fmt.Errorf("servant request unfound events %s", reflect.TypeOf(reqMsg)))
		// 	}

		// 	background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
		// 	background = NewCtxWithServantClientInfo(background, NewClientInfo(ctx, messageEnvelope))
		// 	result, err := evt.(func(context.Context) (proto.Message, error))(background)
		// 	background = FreeCtxWithServantClientInfo(background)
		// 	if err != nil {
		// 		result = &messages.RpcError{Err: err.Error()}
		// 	}

		// 	select {
		// 	case <-background.Done():
		// 		// 已超时，不再需要回复
		// 	default:
		// 		b, _ := messages.MarshalResponseProtobuf(message.SequenceID, result)
		// 		if b != nil {
		// 			ctx.PostMessage(ctx.Self(), b)
		// 		}
		// 	}

		// 	cancel()
		// default:
		// 	panic("Unknown service")
		// }
		// servant_client_offset_label:
		if offset == len(ctx.Message()) {
			break
		}
	}
}

// Register 注册方法映射
func (c *ServantClientConn) Register(p thrift.TProcessor) {
	c.processor = p
}

// func (c *ServantClientConn) Register(key, evt interface{}) {

// 	if evt.(func(context.Context) (proto.Message, error)) == nil {
// 		panic("The input event must be func(context.Context) or " +
// 			"func(context.Context) (proto.Message, error)")
// 	}

// 	c.events[reflect.TypeOf(key)] = evt
// }

func (c *ServantClientConn) Closed(ctx network.Context) {
	c.Servant.clientMutex.Lock()
	delete(c.Servant.clients, network.Key(ctx.Self()))
	c.Servant.clientMutex.Unlock()

	ctxBack := NewCtxWithServantClientInfo(context.Background(), NewClientInfo(ctx, nil))
	c.actor.Closed(ctxBack)
	FreeCtxWithServantClientInfo(ctxBack)
	c.iprot.Close()
	c.oprot.Close()
	// c.recvice.Close()
}

// Ping 客户端主动请求, 这里不处理
func (c *ServantClientConn) Ping(ctx network.Context) {
	ping := messages.NewRpcPingMessage()
	ping.VerifyKey =int64(fastrand.Uint64n(math.MaxUint64))
	b,err :=  messages.MarshalTStruct(context.Background(), c.oprot,ping,1)
	if err != nil{
		return
	}
	b,err = messages.Marshal(b)
	if err != nil{
		return
	}
	ctx.PostMessage(ctx.Self(), b)
}

func (c *ServantClientConn) onRpcPing(ctx network.Context, iprot protocol.IProtocol) {
	// ping := messages.NewRpcPingMessage()
	// ping.Read(context.Background(), iprot)
	// ping.VerifyKey += 1

	// b,err :=  messages.MarshalTStruct(context.Background(), c.oprot,ping,1)
	// if err != nil{
	// 	return
	// }
	// b,err = messages.Marshal(b)
	// if err != nil{
	// 	return
	// }

	// ctx.PostMessage(ctx.Self(), b)
}

func (c *ServantClientConn) onRpcRequest(ctx network.Context, iprot protocol.IProtocol) {
	request := messages.NewRpcRequestMessage()
	request.Read(context.Background(), iprot)

	// 	//TODO: 这里需要抛给并行器
	timeout := int64(request.Timeout) - (time.Now().UnixMilli() - int64(request.ForwardTime))
	// 如果已超时
	if timeout <= 0 {
		return
	}
	rprot := protocol.NewBinaryProtocol()
	defer rprot.Close()
	rprot.Write(request.Message)
	reqMsgName, _, _, err := rprot.ReadMessageBegin(context.Background())
	if err != nil{
		return
	}
	var messageEnvelope *MessageEnvelope
	switch reqMsgName {
	case "prvs.ForwardBundle":
		rs := prvs.NewForwardBundle()
		rs.Read(context.Background(),rprot)
		rprot.Release()
		rprot.Write(rs.Body)
		messageEnvelope = NewMessageEnvelopePool(request.SequenceID, rs.Sender, nil)
	default:
		messageEnvelope = NewMessageEnvelopePool(request.SequenceID, nil, nil)
	}

	background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	background = NewCtxWithServantClientInfo(background, NewClientInfo(ctx, messageEnvelope))
	_, err = c.processor.Process(background, rprot, iprot)
	background = FreeCtxWithServantClientInfo(background)
	if err != nil{
		rprot.Release()
		er := &messages.RpcError{Err: err.Error()}
		rprot.WriteMessageBegin(context.Background(), protocol.MessageName(er), thrift.EXCEPTION, request.SequenceID)
		er.Write(context.Background(), rprot)
		rprot.WriteMessageEnd(context.Background())
		
	 	response := messages.NewRpcResponseMessage()
		response.Result_ = rprot.GetBytes()
		response.SequenceID = request.SequenceID
		rprot.Release()
		rprot.WriteMessageBegin(context.Background(), protocol.MessageName(response), thrift.EXCEPTION, request.SequenceID)
		response.Write(context.Background(),rprot)
		rprot.WriteMessageEnd(context.Background())
		b, err := messages.Marshal(rprot.GetBytes())
		if err != nil{
			return
		}
		ctx.PostMessage(ctx.Self(),b)
	}

	select {
	case <-background.Done():
		// 已超时，不再需要回复
	default:
	}
	cancel()
}

func (c *ServantClientConn) incarnateActor() {
	c.actor = c.Servant.Config.Producer(c)
}
