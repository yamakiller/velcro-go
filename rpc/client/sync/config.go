package sync

import rpcmessage "github.com/yamakiller/velcro-go/rpc/messages"

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
