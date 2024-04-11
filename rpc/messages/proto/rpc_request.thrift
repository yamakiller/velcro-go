namespace go rpc.messages

struct RpcRequestMessage {
    1:i32   SequenceID;
	2:i64 ForwardTime;
	3:i64 Timeout;
	4:binary Message;    
}