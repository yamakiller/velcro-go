package messageproxy

import (
	"context"
	"fmt"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/protocol"
)


func NewMessageProxy(args ...IMessageProxy)IMessageProxy{
	tmp := args[0]
	if len(args) > 1{
		for i:= 1; i < len(args); i++{
			tmp.WithNext(args[i])
			tmp = args[i]
		}
	}
	return args[0]
}

type IMessageProxy interface {
	IMessageProxyNode
	Message(ctx network.Context, msg []byte,timeout int64) error
}

type IMessageProxyStruct interface{
	UnMarshal(msg []byte) error
	Method(ctx network.Context,seqid int32,timeout int64) error
}

func NewRepeatMessageProxy() *RepeatMessageProxy {
	return &RepeatMessageProxy{
		IMessageProxyNode: NewMessageProxyNode(),
		methods:          make(map[string]IMessageProxyStruct),
		iprot: protocol.NewBinaryProtocol(),
	}
}

type RepeatMessageProxy struct {
	IMessageProxyNode
	methods map[string]IMessageProxyStruct
	default_methods IMessageProxyStruct
	iprot protocol.IProtocol
}

func (d *RepeatMessageProxy) Register(key string, proxy IMessageProxyStruct) {
	d.methods[key] = proxy
}

func (d *RepeatMessageProxy) WithDefaultMethod(proxy IMessageProxyStruct){
	d.default_methods = proxy
}

func (d *RepeatMessageProxy) Message(ctx network.Context, msg []byte,timeout int64) error {
	d.iprot.Release()
	d.iprot.Write(msg)
	name, _, seqid, err := d.iprot.ReadMessageBegin(context.Background())
	if err != nil {
		return err
	}
	if info, ok := d.methods[name]; ok {
		if err := info.UnMarshal(msg); err != nil{
			return err
		}
		return info.Method(ctx,seqid,timeout)
	}
	if d.default_methods != nil{
		if err := d.default_methods.UnMarshal(msg); err != nil{
			return err
		}
		return d.default_methods.Method(ctx,seqid,timeout)
	}
	return fmt.Errorf("unkown message %s", name)
}