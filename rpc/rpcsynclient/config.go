package rpcsynclient

import "github.com/yamakiller/velcro-go/rpc/rpcmessage"

type ConnConfig struct {
	MarshalRequest rpcmessage.MarshalRequestFunc
	MarshalMessage rpcmessage.MarshalMessageFunc
	MarshalPing    rpcmessage.MarshalPingFunc

	UnMarshal rpcmessage.UnMarshalFunc
}

func defaultConfig() *ConnConfig {
	return &ConnConfig{
		MarshalRequest: rpcmessage.MarshalRequestProtobuf,
		MarshalMessage: rpcmessage.MarshalMessageProtobuf,
		MarshalPing:    rpcmessage.MarshalPing,
		UnMarshal:      rpcmessage.UnMarshalProtobuf,
	}
}
