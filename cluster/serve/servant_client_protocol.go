package serve

import (
	"context"

	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
)

func  NewServantClientProtocol()*ServantClientProtocol{
	return &ServantClientProtocol{
		BinaryProtocol: protocol.NewBinaryProtocol(),
	}
}

type ServantClientProtocol struct {
	*protocol.BinaryProtocol
}


func (r *ServantClientProtocol) Flush(ctx context.Context) (err error){
	ctxx:= GetServantClientInfo(ctx)
	if ctxx == nil{
		return ctx.Err()
	}
	response := messages.NewRpcResponseMessage()
	response.Result_ = r.GetBytes()
	response.SequenceID = ctxx.SeqId()
	b, err := messages.MarshalTStruct(ctx,r,response,protocol.MessageName(response), ctxx.SeqId())
	if err != nil{
		return err
	}
	b, err = messages.Marshal(b)
	if err != nil{
		return err
	}
	return ctxx.context.PostMessage(ctxx.context.Self(),b)
}