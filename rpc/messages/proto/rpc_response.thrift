namespace go rpc.messages

struct RpcResponseMessage {
    1:i32 SequenceID;
	2:binary Result;
}

struct RpcGuardianResponseMessage {
    1:i32 SequenceID;
	2:binary Result;
}

struct RpcError {
    1:string Err;
}