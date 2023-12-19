package rpcserver

import (
	"github.com/yamakiller/velcro-go/cluster/rpc/rpcmessage"
	"github.com/yamakiller/velcro-go/network"
)

type ConnConfig struct {
	network.Config

	Kleepalive int32

	MarshalResponse rpcmessage.MarshalResponseFunc
	MarshalMessage rpcmessage.MarshalMessageFunc
	MarshalPing    rpcmessage.MarshalPingFunc

	UnMarshal rpcmessage.UnMarshalFunc

	Pool RpcPool
}

func defaultConfig() *ConnConfig {
	return &ConnConfig{
		Kleepalive:     10 * 1000,
		MarshalResponse: rpcmessage.MarshalResponseProtobuf,
		MarshalMessage: rpcmessage.MarshalMessageProtobuf,
		MarshalPing:    rpcmessage.MarshalPing,
		UnMarshal:      rpcmessage.UnMarshalProtobuf,
		Pool: nil,
	}
}