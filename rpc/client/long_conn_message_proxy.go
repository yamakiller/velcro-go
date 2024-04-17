package client

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
)

func NewRpcPingMessageProxy(c *LongConn)*RpcPingMessageProxy{
	return &RpcPingMessageProxy{
		message: messages.NewRpcPingMessage(),
		iprot: protocol.NewBinaryProtocol(),
		c:c,
	}
}
type RpcPingMessageProxy struct {
	message *messages.RpcPingMessage
	iprot   protocol.IProtocol
	c *LongConn
}

func (slf *RpcPingMessageProxy) UnMarshal(msg []byte) error {
	slf.iprot.Release()
	slf.iprot.Write(msg)
	_, _, err := messages.UnMarshalTStruct(context.Background(), slf.iprot, slf.message)
	if err != nil {
		return err
	}
	return nil
}

func (slf *RpcPingMessageProxy) Method(ctx network.Context, seqId int32, timeout int64) error {

	slf.message.VerifyKey += 1
	b, err := messages.MarshalTStruct(context.Background(), slf.iprot, slf.message, protocol.MessageName(slf.message), seqId)
	if err != nil {
		return err
	}
	
	b,err = messages.Marshal(b)
	if err != nil{
		return err
	}
	_, err = slf.c.conn.Write(b)
	return err
}

func NewRpcResponseMessageProxy(c *LongConn) *RpcResponseMessageProxy{
	return &RpcResponseMessageProxy{
		message: messages.NewRpcResponseMessage(),
		iprot: protocol.NewBinaryProtocol(),
		c:c,
	}
}
type RpcResponseMessageProxy struct{
	message *messages.RpcResponseMessage
	iprot   protocol.IProtocol
	c *LongConn
}
func (slf *RpcResponseMessageProxy) UnMarshal(msg []byte) error {
	slf.iprot.Release()
	slf.iprot.Write(msg)
	_, _, err := messages.UnMarshalTStruct(context.Background(), slf.iprot, slf.message)
	if err != nil {
		return err
	}
	return nil
}


func (slf *RpcResponseMessageProxy) Method(ctx network.Context, seqId int32, timeout int64) error {
	response := &messages.RpcGuardianResponseMessage{SequenceID:  slf.message.SequenceID,Result_:  slf.message.Result_} 
	msg,_:= messages.MarshalTStruct(context.Background(),slf.iprot,response,protocol.MessageName(response),response.SequenceID)
	slf.c.mailbox <- msg
	return nil
}

func NewRpcGuardianResponseMessageProxy(c *LongConn) *RpcGuardianResponseMessageProxy{
	return &RpcGuardianResponseMessageProxy{
		message: messages.NewRpcGuardianResponseMessage(),
		err: messages.NewRpcError(),
		iprot: protocol.NewBinaryProtocol(),
		c:c,
	}
}
type RpcGuardianResponseMessageProxy struct{
	message *messages.RpcGuardianResponseMessage
	err 	*messages.RpcError
	iprot   protocol.IProtocol
	c *LongConn
}

func (slf *RpcGuardianResponseMessageProxy) UnMarshal(msg []byte) error {
	slf.iprot.Release()
	slf.iprot.Write(msg)
	_, _, err := messages.UnMarshalTStruct(context.Background(), slf.iprot, slf.message)
	if err != nil {
		return err
	}
	if slf.message.Result_ != nil{
		slf.iprot.Release()
		slf.iprot.Write(slf.message.Result_)
		
		if name,_,_ := messages.UnMarshalTStruct(context.Background(), slf.iprot, slf.err); name != protocol.MessageName(slf.err){
			slf.err.Err = ""
		}
	}
	return nil
}
func (slf *RpcGuardianResponseMessageProxy) Method(ctx network.Context, seqId int32, timeout int64) error {
	future := slf.c.getFuture(slf.message.SequenceID)
	if future == nil {
		return nil
	}

	future.cond.L.Lock()
	if future.done {
		future.cond.L.Unlock()
		return nil
	}
	future.result = slf.message.Result_
	future.err = nil
	if slf.err.Err != ""{
		future.err = errors.New(slf.err.Err)
	}

	future.done = true

	tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
	if tp != nil {
		tp.Stop()
	}
	future.cond.Signal()
	future.cond.L.Unlock()
	return nil
}