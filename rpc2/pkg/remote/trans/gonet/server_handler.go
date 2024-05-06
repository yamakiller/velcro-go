package gonet

import (
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote/trans"
)

type svrTransHandlerFactory struct{}

func NewSvrTransHandlerFactory() remote.ServerTransHandlerFactory {
	return &svrTransHandlerFactory{}
}

func (f *svrTransHandlerFactory) NewTransHandler(opt *remote.ServerOption) (remote.ServerTransHandler, error) {
	return newSvrTransHandler(opt)
}

func newSvrTransHandler(opt *remote.ServerOption) (remote.ServerTransHandler, error) {
	return trans.NewDefaultSvrTransHandler(opt, NewGonetExtAddon())
}
