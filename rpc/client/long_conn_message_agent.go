package client

/*func NewMessageAgent(conn *LongConn)messageagent.IMessageAgent{
	repeat := messageagent.NewRepeatMessageAgent()
	repeat.Register(protocol.MessageName(&messages.RpcResponseMessage{}),NewRpcResponseMessageAgent(conn))
	repeat.Register(protocol.MessageName(&messages.RpcPingMessage{}),NewRpcPingMessageAgent(conn))
	return repeat
}
func NewGuardianMessageAgent(conn *LongConn)messageagent.IMessageAgent{
	repeat := messageagent.NewRepeatMessageAgent()
	repeat.Register(protocol.MessageName(&messages.RpcGuardianResponseMessage{}),NewRpcGuardianResponseMessageAgent(conn))
	return repeat
}
func NewRpcPingMessageAgent(c *LongConn)*RpcPingMessageAgent{
	return &RpcPingMessageAgent{
		message: messages.NewRpcPingMessage(),
		iprot: protocol.NewBinaryProtocol(),
		c:c,
	}
}
type RpcPingMessageAgent struct {
	message *messages.RpcPingMessage
	iprot   protocol.IProtocol
	c *LongConn
}

func (slf *RpcPingMessageAgent) UnMarshal(msg []byte) error {
	slf.iprot.Release()
	slf.iprot.Write(msg)
	_, _, err := messages.UnMarshalTStruct(context.Background(), slf.iprot, slf.message)
	if err != nil {
		return err
	}
	return nil
}

func (slf *RpcPingMessageAgent) Method(ctx network.Context, seqId int32, timeout int64) error {

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

func NewRpcResponseMessageAgent(c *LongConn) *RpcResponseMessageAgent{
	return &RpcResponseMessageAgent{
		message: messages.NewRpcResponseMessage(),
		iprot: protocol.NewBinaryProtocol(),
		c:c,
	}
}
type RpcResponseMessageAgent struct{
	message *messages.RpcResponseMessage
	iprot   protocol.IProtocol
	c *LongConn
}
func (slf *RpcResponseMessageAgent) UnMarshal(msg []byte) error {
	slf.iprot.Release()
	slf.iprot.Write(msg)
	_, _, err := messages.UnMarshalTStruct(context.Background(), slf.iprot, slf.message)
	if err != nil {
		return err
	}
	return nil
}


func (slf *RpcResponseMessageAgent) Method(ctx network.Context, seqId int32, timeout int64) error {
	response := &messages.RpcGuardianResponseMessage{SequenceID:  slf.message.SequenceID,Result_:  slf.message.Result_}
	msg,_:= messages.MarshalTStruct(context.Background(),slf.iprot,response,protocol.MessageName(response),response.SequenceID)
	slf.c.mailbox <- msg
	return nil
}

func NewRpcGuardianResponseMessageAgent(c *LongConn) *RpcGuardianResponseMessageAgent{
	return &RpcGuardianResponseMessageAgent{
		message: messages.NewRpcGuardianResponseMessage(),
		err: messages.NewRpcError(),
		iprot: protocol.NewBinaryProtocol(),
		c:c,
	}
}
type RpcGuardianResponseMessageAgent struct{
	message *messages.RpcGuardianResponseMessage
	err 	*messages.RpcError
	iprot   protocol.IProtocol
	c *LongConn
}

func (slf *RpcGuardianResponseMessageAgent) UnMarshal(msg []byte) error {
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
func (slf *RpcGuardianResponseMessageAgent) Method(ctx network.Context, seqId int32, timeout int64) error {
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
}*/
