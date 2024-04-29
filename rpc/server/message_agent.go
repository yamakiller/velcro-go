package server

/*import (
	"context"
	"time"

	messageagent "github.com/yamakiller/velcro-go/cluster/agent/message"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
)

func NewRpcClientMessageAgent(conn *RpcClientConn)messageagent.IMessageAgent{
	repeat := messageagent.NewRepeatMessageAgent()
	repeat.Register(protocol.MessageName(&messages.RpcPingMessage{}),NewRpcPingMessageAgent())
	repeat.Register(protocol.MessageName(&messages.RpcRequestMessage{}),NewRpcRequestMessageAgent(conn))
	repeat.Register(protocol.MessageName(&messages.RpcResponseMessage{}),NewRpcResponseMessageAgent(conn))
	return repeat
}

func NewRpcPingMessageAgent()*RpcPingMessageAgent{
	return &RpcPingMessageAgent{
		message: messages.NewRpcPingMessage(),
		iprot: protocol.NewBinaryProtocol(),
	}
}

type RpcPingMessageAgent struct {
	message *messages.RpcPingMessage
	iprot protocol.IProtocol
}

func (slf *RpcPingMessageAgent) UnMarshal(msg []byte) error {
	slf.iprot.Release()
	slf.iprot.Write(msg)
	_,_, err := messages.UnMarshalTStruct(context.Background(),slf.iprot,slf.message)
	if err != nil{
		return err
	}
	return nil
}

func (slf *RpcPingMessageAgent) Method(ctx network.Context, seqId int32, timeout int64) error {

	slf.message.VerifyKey += 1
	b, err:= messages.MarshalTStruct(context.Background(),slf.iprot, slf.message,protocol.MessageName(slf.message),seqId)
	if err != nil{
		return err
	}
	ctx.PostMessage(ctx.Self(), b)

	return nil
}


func NewRpcRequestMessageAgent(rcc *RpcClientConn)*RpcRequestMessageAgent{
	return &RpcRequestMessageAgent{
		message: messages.NewRpcRequestMessage(),
		iprot: protocol.NewBinaryProtocol(),
		rcc:rcc,
	}
}

type RpcRequestMessageAgent struct {
	message *messages.RpcRequestMessage
	iprot protocol.IProtocol
	rcc *RpcClientConn
}

func (slf *RpcRequestMessageAgent) UnMarshal(msg []byte) error {
	slf.iprot.Release()
	slf.iprot.Write(msg)
	_,_, err := messages.UnMarshalTStruct(context.Background(),slf.iprot,slf.message)
	if err != nil{
		return err
	}
	slf.iprot.Release()
	slf.iprot.Write(slf.message.Message)
	return nil
}

func (slf *RpcRequestMessageAgent) Method(ctx network.Context, seqId int32, _ int64) error {

	timeout := int64(slf.message.Timeout) - (time.Now().UnixMilli() - int64(slf.message.ForwardTime))
	// 如果已超时
	if timeout <= 0 {
		return nil
	}
	// 设置超时器
	background, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	ctxx :=NewCtxWithRpcClientContext(background, &RpcClientContext{sequenceID: slf.message.SequenceID,
		background: background,
		context:    ctx})

		slf.rcc.processor.Process(ctxx, slf.iprot, slf.rcc.oprot)
	FreeCtxWithRpcClientContext(ctxx)
	// 判断是否超时
	select {
	case <-background.Done():
	default:
	}

	cancel()
	return nil
}

func NewRpcResponseMessageAgent(rcc *RpcClientConn)*RpcResponseMessageAgent{
	return &RpcResponseMessageAgent{
		message: messages.NewRpcResponseMessage(),
		iprot: protocol.NewBinaryProtocol(),
		rcc:rcc,
	}
}

type RpcResponseMessageAgent struct {
	message *messages.RpcResponseMessage
	iprot protocol.IProtocol
	rcc *RpcClientConn
}

func (slf *RpcResponseMessageAgent) UnMarshal(msg []byte) error {
	slf.iprot.Release()
	slf.iprot.Write(msg)
	_,_, err := messages.UnMarshalTStruct(context.Background(),slf.iprot,slf.message)
	if err != nil{
		return err
	}
	slf.iprot.Release()
	slf.iprot.Write(slf.message.Result_)
	return nil
}

func (slf *RpcResponseMessageAgent) Method(ctx network.Context, seqId int32, timeout int64) error {
	background := context.Background()
	ctxx :=NewCtxWithRpcClientContext(background, &RpcClientContext{sequenceID: slf.message.SequenceID,
		background: background,
		context:    ctx})
		slf.rcc.processor.Process(ctxx, slf.iprot,slf.rcc. oprot)
	FreeCtxWithRpcClientContext(ctxx)
	return nil
}*/
