package asyn

import rpcmessage "github.com/yamakiller/velcro-go/rpc/messages"

type ConnConfig struct {
	Kleepalive int32

	MarshalRequest rpcmessage.MarshalRequestFunc
	MarshalMessage rpcmessage.MarshalMessageFunc
	MarshalPing    rpcmessage.MarshalPingFunc

	UnMarshal rpcmessage.UnMarshalFunc

	Connected ConnectedFunc
	Receive   ReceiveFunc
	Closed    ClosedFunc
}

func defaultConfig() *ConnConfig {
	return &ConnConfig{
		Kleepalive:     10 * 1000,
		MarshalRequest: rpcmessage.MarshalRequestProtobuf,
		MarshalMessage: rpcmessage.MarshalMessageProtobuf,
		MarshalPing:    rpcmessage.MarshalPing,
		UnMarshal:      rpcmessage.UnMarshalProtobuf,
	}
}
