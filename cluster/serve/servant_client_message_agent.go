package serve

import (
	"context"
	"time"

	messageagent "github.com/yamakiller/velcro-go/cluster/agent/message"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
)


func NewServantMessageAgent(conn *ServantClientConn)messageagent.IMessageAgent{
	requset := NewRpcRequestMessageAgent()
	requset.Register(protocol.MessageName(&prvs.ForwardBundle{}),NewForwardBundleMessageAgent(conn))
	requset.WithDefaultMethod(NewDefaultRpcRequestMessageAgent(conn))

	repeat := messageagent.NewRepeatMessageAgent()
	repeat.Register(protocol.MessageName(&messages.RpcRequestMessage{}),requset)
	repeat.Register(protocol.MessageName(&messages.RpcPingMessage{}),NewRpcPingMessageAgent())
	return repeat
}

func NewRpcPingMessageAgent() *RpcPingMessageAgent {
	return &RpcPingMessageAgent{
		message:           messages.NewRpcPingMessage(),
		iprot:             protocol.NewBinaryProtocol(),
	}
}

type RpcPingMessageAgent struct {
	message *messages.RpcPingMessage
	iprot   protocol.IProtocol
}

func (rpmp *RpcPingMessageAgent) UnMarshal(msg []byte) error {
	rpmp.iprot.Release()
	rpmp.iprot.Write(msg)
	_,_, err := messages.UnMarshalTStruct(context.Background(),rpmp.iprot,rpmp.message)
	if err != nil{
		return err
	}
	return nil
}

func (rpmp *RpcPingMessageAgent) Method(ctx network.Context, seqId int32, timeout int64) error {
	return nil
}

func NewRpcRequestMessageAgent() *RpcRequestMessageAgent {
	return &RpcRequestMessageAgent{
		message:           messages.NewRpcRequestMessage(),
		repeat:            messageagent.NewRepeatMessageAgent(),
		iprot:             protocol.NewBinaryProtocol(),
	}
}

type RpcRequestMessageAgent struct {
	message *messages.RpcRequestMessage
	repeat  *messageagent.RepeatMessageAgent
	iprot   protocol.IProtocol
}

func (rrmp *RpcRequestMessageAgent) Register(key string, agent messageagent.IMessageAgentStruct){
	rrmp.repeat.Register(key,agent)
}
func (rrmp *RpcRequestMessageAgent) WithDefaultMethod(agent messageagent.IMessageAgentStruct){
	rrmp.repeat.WithDefaultMethod(agent)
}

func (rrmp *RpcRequestMessageAgent) UnMarshal(msg []byte) error {
	rrmp.iprot.Release()
	rrmp.iprot.Write(msg)
	_,_, err := messages.UnMarshalTStruct(context.Background(),rrmp.iprot,rrmp.message)
	if err != nil{
		return err
	}
	return nil
}

func (rrmp *RpcRequestMessageAgent) Method(ctx network.Context, seqId int32, _ int64) error {

	// 	//TODO: 这里需要抛给并行器
	timeout := int64(rrmp.message.Timeout) - (time.Now().UnixMilli() - int64(rrmp.message.ForwardTime))
	// 如果已超时
	if timeout <= 0 {
		return nil
	}

	return rrmp.repeat.Message(ctx, rrmp.message.Message, timeout)
}


func NewForwardBundleMessageAgent(conn  *ServantClientConn)*ForwardBundleMessageAgent{
	return &ForwardBundleMessageAgent{
		message: prvs.NewForwardBundle(),
		iprot: protocol.NewBinaryProtocol(),
		conn: conn,
	}
}
type ForwardBundleMessageAgent struct{
	message *prvs.ForwardBundle
	iprot protocol.IProtocol
	conn  *ServantClientConn
}

func (fbmp *ForwardBundleMessageAgent) UnMarshal(msg []byte) error{
	fbmp.iprot.Release()
	fbmp.iprot.Write(msg)
	_,_, err := messages.UnMarshalTStruct(context.Background(),fbmp.iprot,fbmp.message)
	if err != nil{
		return err
	}
	fbmp.iprot.Release()
	fbmp.iprot.Write(fbmp.message.Body)
	return nil
}

func (fbmp *ForwardBundleMessageAgent) Method(ctx network.Context, seqId int32,timeout int64) error{

	messageEnvelope := NewMessageEnvelopePool(seqId, fbmp.message.Sender, nil)
	background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	background = NewCtxWithServantClientInfo(background, NewClientInfo(ctx, messageEnvelope))
	_, err := fbmp.conn.processor.Process(background, fbmp.iprot, fbmp.conn.oprot)
	background = FreeCtxWithServantClientInfo(background)
	if err != nil{
		er := &messages.RpcError{Err: err.Error()}
		erdata,_ :=messages.MarshalTStruct(context.Background(),fbmp.iprot,er,protocol.MessageName(er),seqId)
	 	response := messages.NewRpcResponseMessage()
		response.Result_ = erdata
		response.SequenceID = seqId
		resdata,_ :=messages.MarshalTStruct(context.Background(),fbmp.iprot,response,protocol.MessageName(response),seqId)
		b, err := messages.Marshal(resdata)
		if err != nil{
			return err
		}
		ctx.PostMessage(ctx.Self(),b)
	}

	select {
	case <-background.Done():
		// 已超时，不再需要回复
	default:
	}
	cancel()
	return nil
}

func NewDefaultRpcRequestMessageAgent(conn  *ServantClientConn)*DefaultRpcRequestMessageAgent{
	return &DefaultRpcRequestMessageAgent{
		iprot: protocol.NewBinaryProtocol(),
		conn:conn,
	}
}
type DefaultRpcRequestMessageAgent struct{
	iprot protocol.IProtocol
	conn  *ServantClientConn

}

func (drrmp *DefaultRpcRequestMessageAgent) UnMarshal(msg []byte) error{
	drrmp.iprot.Release()
	drrmp.iprot.Write(msg)
	return nil
}
func (drrmp *DefaultRpcRequestMessageAgent) Method(ctx network.Context, seqId int32,timeout int64) error{
	messageEnvelope := NewMessageEnvelopePool(seqId, nil, nil)
	background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	background = NewCtxWithServantClientInfo(background, NewClientInfo(ctx, messageEnvelope))

	_, err := drrmp.conn.processor.Process(background, drrmp.iprot, drrmp.conn.oprot)
	background = FreeCtxWithServantClientInfo(background)
	if err != nil{
		er := &messages.RpcError{Err: err.Error()}
		erdata,_ :=messages.MarshalTStruct(context.Background(),drrmp.iprot,er,protocol.MessageName(er),seqId)
	 	response := messages.NewRpcResponseMessage()
		response.Result_ = erdata
		response.SequenceID = seqId
		resdata,_ :=messages.MarshalTStruct(context.Background(),drrmp.iprot,response,protocol.MessageName(response),seqId)
		b, err := messages.Marshal(resdata)
		if err != nil{
			return err
		}
		ctx.PostMessage(ctx.Self(),b)
	}

	select {
	case <-background.Done():
		// 已超时，不再需要回复
	default:
	}
	cancel()
	return nil
}