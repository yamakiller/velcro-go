package serve

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"google.golang.org/protobuf/proto"
)

type ServantClientContext struct {
	SequenceID int32
	Sender     *network.ClientID
	Background context.Context
	Context    network.Context
	Message    interface{}
}

type ServantClientConn struct {
	Servant *Servant
	vaddr   string
	actor   ServantClientActor
	recvice *circbuf.RingBuffer // 接收缓冲区
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
		n, err := c.recvice.Write(ctx.Message()[offset:])
		offset += n

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
				panic(err)
			}

			timeout := int64(message.Timeout) - (time.Now().UnixMilli() - int64(message.ForwardTime))
			// 如果已超时
			if timeout <= 0 {
				goto servant_client_offset_label
			}

			var sender *network.ClientID = nil
			switch rs := reqMsg.(type) {
			case *protocols.RegisterRequest:
				c.vaddr = rs.Vaddr
				bucket := c.Servant.vaddrs.getBucket(c.vaddr)
				bucket.SetIfAbsent(c.vaddr, ctx.Self())

				b, _ := messages.MarshalResponseProtobuf(message.SequenceID, &protocols.RegisterResponse{})
				ctx.PostMessage(ctx.Self(), b)
				goto servant_client_offset_label
			case *protocols.Forward:
				reqMsg, err = rs.Msg.UnmarshalNew()
				if err != nil {
					panic(err)
				}
				sender = rs.Sender
			default:
			}
			// 设置超时器
			background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
			evt, ok := c.events[reflect.TypeOf(reqMsg)]
			if !ok || evt.(func(*ServantClientContext) (proto.Message, error)) == nil {
				panic(fmt.Errorf("servant request unfound events %s", reflect.TypeOf(reqMsg)))
			}
			result, err := evt.(func(*ServantClientContext) (proto.Message, error))(&ServantClientContext{
				SequenceID: message.SequenceID,
				Sender:     sender,
				Background: background,
				Context:    ctx,
				Message:    reqMsg,
			})

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
		case *messages.RpcMsgMessage:
			postMsg, err := message.Message.UnmarshalNew()
			if err != nil {
				goto servant_client_offset_label
			}

			var sender *network.ClientID = nil
			switch rs := postMsg.(type) {
			case *protocols.Forward:
				postMsg, err = rs.Msg.UnmarshalNew()
				if err != nil {
					panic(err)
				}
				sender = rs.Sender
			default:
			}

			evt, ok := c.events[reflect.TypeOf(postMsg)]
			if !ok || evt.(func(*ServantClientContext)) == nil {
				panic(fmt.Errorf("servant post unfound events %s", reflect.TypeOf(postMsg)))
			}

			evt.(func(*ServantClientContext))(&ServantClientContext{
				SequenceID: message.SequenceID,
				Sender:     sender,
				Background: nil,
				Context:    ctx,
				Message:    postMsg,
			})
		default:
			panic("Unknown service")
		}
	servant_client_offset_label:
		if offset == len(ctx.Message()) {
			break
		}
	}
}

// Register 注册方法映射
func (c *ServantClientConn) Register(key, evt interface{}) {

	if evt.(func(*ServantClientContext)) == nil &&
		evt.(func(*ServantClientContext) (proto.Message, error)) == nil {
		panic("The input event must be func(*ServantClientContext) or " +
			"func(*ServantClientContext) (proto.Message, error)")
	}

	c.events[reflect.TypeOf(key)] = evt
}

func (c *ServantClientConn) Closed(ctx network.Context) {
	if c.vaddr != "" {
		bucket := c.Servant.vaddrs.getBucket(c.vaddr)
		_, _ = bucket.Pop(c.vaddr)
	}

	c.Servant.clientMutex.Lock()
	delete(c.Servant.clients, network.Key(ctx.Self()))
	c.Servant.clientMutex.Unlock()

	c.actor.Closed(&ServantClientContext{Context: ctx})
	c.recvice = nil
}

// Ping 客户端主动请求, 这里不处理
func (c *ServantClientConn) Ping(ctx network.Context) {

}

func (c *ServantClientConn) onRpcPing(ctx network.Context, message *messages.RpcPingMessage) {
	b, err := messages.MarshalPingProtobuf(message.VerifyKey + 1)
	if err != nil {
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), b)
}

func (c *ServantClientConn) incarnateActor() {
	c.actor = c.Servant.Config.Producer(c)
}
