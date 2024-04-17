package serve

import (
	"context"
	"time"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/proxy/messageproxy"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
)

func NewRpcPingMessageProxy() *RpcPingMessageProxy {
	return &RpcPingMessageProxy{
		IMessageProxyNode: messageproxy.NewMessageProxyNode(),
		message:           messages.NewRpcPingMessage(),
		iprot:             protocol.NewBinaryProtocol(),
	}
}

type RpcPingMessageProxy struct {
	messageproxy.IMessageProxyNode
	message *messages.RpcPingMessage
	iprot   protocol.IProtocol
}

func (rpmp *RpcPingMessageProxy) UnMarshal(msg []byte) error {
	rpmp.iprot.Release()
	rpmp.iprot.Write(msg)
	_,_, err := messages.UnMarshalTStruct(context.Background(),rpmp.iprot,rpmp.message)
	if err != nil{
		return err
	}
	return nil
}

func (rpmp *RpcPingMessageProxy) Method(ctx network.Context, seqId int32, timeout int64) error {
	return nil
}

func NewRpcRequestMessageProxy() *RpcRequestMessageProxy {
	return &RpcRequestMessageProxy{
		IMessageProxyNode: messageproxy.NewMessageProxyNode(),
		message:           messages.NewRpcRequestMessage(),
		repeat:            messageproxy.NewRepeatMessageProxy(),
		iprot:             protocol.NewBinaryProtocol(),
	}
}

type RpcRequestMessageProxy struct {
	messageproxy.IMessageProxyNode
	message *messages.RpcRequestMessage
	repeat  *messageproxy.RepeatMessageProxy
	iprot   protocol.IProtocol
}

func (rrmp *RpcRequestMessageProxy) Register(key string, proxy messageproxy.IMessageProxyStruct){
	rrmp.repeat.Register(key,proxy)
}
func (rrmp *RpcRequestMessageProxy) WithDefaultMethod(proxy messageproxy.IMessageProxyStruct){
	rrmp.repeat.WithDefaultMethod(proxy)
}

func (rrmp *RpcRequestMessageProxy) UnMarshal(msg []byte) error {
	rrmp.iprot.Release()
	rrmp.iprot.Write(msg)
	_,_, err := messages.UnMarshalTStruct(context.Background(),rrmp.iprot,rrmp.message)
	if err != nil{
		return err
	}
	return nil
}

func (rrmp *RpcRequestMessageProxy) Method(ctx network.Context, seqId int32, _ int64) error {

	// 	//TODO: 这里需要抛给并行器
	timeout := int64(rrmp.message.Timeout) - (time.Now().UnixMilli() - int64(rrmp.message.ForwardTime))
	// 如果已超时
	if timeout <= 0 {
		return nil
	}

	return rrmp.repeat.Message(ctx, rrmp.message.Message, timeout)
}


func NewForwardBundleMessageProxy(conn  *ServantClientConn)*ForwardBundleMessageProxy{
	return &ForwardBundleMessageProxy{
		message: prvs.NewForwardBundle(),
		iprot: protocol.NewBinaryProtocol(),
		conn: conn,
	}
}
type ForwardBundleMessageProxy struct{
	message *prvs.ForwardBundle
	iprot protocol.IProtocol
	conn  *ServantClientConn
}

func (fbmp *ForwardBundleMessageProxy) UnMarshal(msg []byte) error{
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

func (fbmp *ForwardBundleMessageProxy) Method(ctx network.Context, seqId int32,timeout int64) error{

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

func NewDefaultRpcRequestMessageProxy(conn  *ServantClientConn)*DefaultRpcRequestMessageProxy{
	return &DefaultRpcRequestMessageProxy{
		iprot: protocol.NewBinaryProtocol(),
		conn:conn,
	}
}
type DefaultRpcRequestMessageProxy struct{
	iprot protocol.IProtocol
	conn  *ServantClientConn

}

func (drrmp *DefaultRpcRequestMessageProxy) UnMarshal(msg []byte) error{
	drrmp.iprot.Release()
	drrmp.iprot.Write(msg)
	return nil
}
func (drrmp *DefaultRpcRequestMessageProxy) Method(ctx network.Context, seqId int32,timeout int64) error{
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