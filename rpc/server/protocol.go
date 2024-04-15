package server

import (
	"context"

	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
)

func NewRpcContextProtocol()*RpcContextProtocol{
return &RpcContextProtocol{
	BinaryProtocol: protocol.NewBinaryProtocol(),
}

}
type RpcContextProtocol struct{
	*protocol.BinaryProtocol
}
func (r *RpcContextProtocol) Flush(ctx context.Context) error{
	ctxx := GetRpcClientContext(ctx)
	if ctxx == nil {
		return nil
	}
	
	response := messages.NewRpcResponseMessage()
	response.Result_ = r.GetBytes()
	response.SequenceID = ctxx.SequenceID()
	b, err := messages.MarshalTStruct(ctx,r,response,protocol.MessageName(response), ctxx.SequenceID())
	if err != nil{
		return err
	}
	return ctxx.context.PostMessage(ctxx.context.Self(),b)
}