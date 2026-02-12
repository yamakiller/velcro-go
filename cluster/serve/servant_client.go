package serve

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
)

// Background context.Context

// ServantClientContext

type ServantClientConn struct {
	Servant *Servant
	actor   ServantClientActor
	recvice *circbuf.LinkBuffer
	events  map[interface{}]interface{}
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
		n, err := c.recvice.WriteBinary(ctx.Message()[offset:])
		offset += n
		if err := c.recvice.Flush(); err != nil {
			ctx.Close(ctx.Self())
			return
		}

		_, msg, msgErr := messages.UnMarshalProtobuf(c.recvice)
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
			goto servant_client_offset_label
		}

		switch message := msg.(type) {
		case *messages.RpcPingMessage:
			c.onRpcPing(ctx, message)
		case *messages.RpcRequestMessage:
			reqMsg, err := message.Message.UnmarshalNew()
			if err != nil {
				vlog.Error(fmt.Sprintf("Unknown Message %v %v %v", ctx.Self().String(), message, err.Error()))
				return
			}

			//TODO: 这里需要抛给并行器
			timeout := int64(message.Timeout) - (time.Now().UnixMilli() - int64(message.ForwardTime))
			// 如果已超时
			if timeout <= 0 {
				goto servant_client_offset_label
			}
			msgType, err := protoregistry.GlobalTypes.FindMessageByName(reqMsg.(*anypb.Any).MessageName())
			if err != nil {
				vlog.Error(fmt.Sprintf("Unknown MessageName %v %v %v", ctx.Self().String(), reqMsg, err.Error()))
				return
			}
			m := msgType.New().Interface()
			proto.Unmarshal(reqMsg.(*anypb.Any).GetValue(), m)
			reqMsg = m

			var messageEnvelope *MessageEnvelope
			switch rs := reqMsg.(type) {
			case *prvs.ForwardBundle:
				reqMsg, err = rs.Body.UnmarshalNew()
				if err != nil {
					vlog.Error(fmt.Sprintf("Unknown Body %v %v %v", ctx.Self().String(), rs.Body, err.Error()))
					return
				}
				messageEnvelope = NewMessageEnvelopePool(message.SequenceID, rs.Sender, reqMsg)
			default:
				messageEnvelope = NewMessageEnvelopePool(message.SequenceID, nil, reqMsg)
			}

			evt, ok := c.events[reflect.TypeOf(reqMsg)]
			if !ok || evt.(func(context.Context) (proto.Message, error)) == nil {
				vlog.Error(fmt.Sprintf("servant request unfound events %v %v", ctx.Self().String(), reflect.TypeOf(reqMsg)))
				return
			}

			background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
			background = NewCtxWithServantClientInfo(background, NewClientInfo(ctx, messageEnvelope))
			result, err := evt.(func(context.Context) (proto.Message, error))(background)
			background = FreeCtxWithServantClientInfo(background)
			if err != nil {
				result = &messages.RpcError{Err: err.Error()}
			}

			select {
			case <-background.Done():
				// 已超时，不再需要回复
			default:
				b, _ := messages.MarshalResponseProtobuf(message.SequenceID, result)
				if b != nil {
					ctx.PostMessage(ctx.Self(), b)
				}
			}

			cancel()
		default:
			vlog.Error(fmt.Sprintf("Unknown service %v  msg %v", ctx.Self().String(), msg))
			return
		}
	servant_client_offset_label:
		if offset == len(ctx.Message()) {
			break
		}
	}
}

// Register 注册方法映射
func (c *ServantClientConn) Register(key, evt interface{}) {

	if evt.(func(context.Context) (proto.Message, error)) == nil {
		panic("The input event must be func(context.Context) or " +
			"func(context.Context) (proto.Message, error)")
	}

	c.events[reflect.TypeOf(key)] = evt
}

func (c *ServantClientConn) Closed(ctx network.Context) {
	c.Servant.clientMutex.Lock()
	delete(c.Servant.clients, network.Key(ctx.Self()))
	c.Servant.clientMutex.Unlock()

	ctxBack := NewCtxWithServantClientInfo(context.Background(), NewClientInfo(ctx, nil))
	c.actor.Closed(ctxBack)
	FreeCtxWithServantClientInfo(ctxBack)

	c.recvice.Close()
}

// Ping 客户端主动请求, 这里不处理
func (c *ServantClientConn) Ping(ctx network.Context) {
	b, err := messages.MarshalPingProtobuf(fastrand.Uint64n(math.MaxUint64))
	if err != nil {
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), b)
}

func (c *ServantClientConn) onRpcPing(ctx network.Context, message *messages.RpcPingMessage) {
	// b, err := messages.MarshalPingProtobuf(message.VerifyKey + 1)
	// if err != nil {
	// 	ctx.Close(ctx.Self())
	// 	return
	// }

	// ctx.PostMessage(ctx.Self(), b)
}

func (c *ServantClientConn) incarnateActor() {
	c.actor = c.Servant.Config.Producer(c)
}
