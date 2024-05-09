package server

import (
	"github.com/yamakiller/velcro-go/rpc/pkg/remote"
	"github.com/yamakiller/velcro-go/utils/velcronet"
)

func newServerRemoteOption(*remote.ServerOption) {
	return &remote.ServerOption{
		TransServerFactory:    velcronet.NewTransServerFactory(),
		SvrHandlerFactory:     detection.NewSvrTransHandlerFactory(velcronet.NewSvrTransHandlerFactory()),
		Codec:                 codec.NewDefaultCode(),
		Address:               defaultAddress,
		ExitWaitTime:          defaultExitWaitTime,
		MaxConnectionIdleTime: defaultConnectionIdleTime,
		AcceptFailedDelayTime: defaultAcceptFailedDelayTime,
	}
}
