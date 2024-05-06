package codec

import (
	"errors"
	"fmt"

	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
)

const (
	// FrontMask is used in protocol sniffing.
	FrontMask = 0x0000ffff
)

// UpdateMsgType is used to set method msgType
func UpdateMsgType(msgType uint32, message remote.Message) error {
	t := message.MessageType()
	if t == remote.MessageType(msgType) {
		return nil
	}
	message.SetMessageType(remote.MessageType(msgType))
	return nil
}

// SetOrCheckSeqID is used to set method seqID to invocation.
func SetOrCheckSeqID(seqID int32, message remote.Message) error {
	ri := message.RPCInfo()
	ink := ri.Invocation()
	callMethodSeqID := ink.SeqID()
	if callMethodSeqID == 0 {
		return fmt.Errorf("method seqid that receive is 0")
	}
	if callMethodSeqID == seqID {
		return nil
	}
	// the server's callMethodSeqID may not be empty if RPCInfo is based on connection multiplexing
	// for the server side callMethodSeqID ! = SeqID is normal
	if message.RPCRole() == remote.Client {
		return fmt.Errorf("wrong method SeqID, expect=%d, actual=%d", callMethodSeqID, seqID)
	}

	if ink, ok := ink.(rpcinfo.InvocationSetter); ok {
		ink.SetSeqID(seqID)
	} else {
		return errors.New("the interface Invocation doesn't implement InvocationSetter")
	}
	return nil
}

// SetOrCheckMethodName is used to set method name to invocation.
func SetOrCheckMethodName(methodName string, message remote.Message) error {
	ri := message.RPCInfo()
	ink := ri.Invocation()
	callMethodName := ink.MethodName()
	if methodName == "" {
		return fmt.Errorf("method name that receive is empty")
	}
	if callMethodName == methodName {
		return nil
	}
	// the server's callMethodName may not be empty if RPCInfo is based on connection multiplexing
	// for the server side callMethodName ! = methodName is normal
	if message.RPCRole() == remote.Client {
		return fmt.Errorf("wrong method name, expect=%s, actual=%s", callMethodName, methodName)
	}
	svcInfo := message.ServiceInfo()
	if ink, ok := ink.(rpcinfo.InvocationSetter); ok {
		ink.SetMethodName(methodName)
		//ink.SetPackageName(svcInfo.GetPackageName())
		ink.SetServiceName(svcInfo.ServiceName)
	} else {
		return errors.New("the interface Invocation doesn't implement InvocationSetter")
	}
	if mt := svcInfo.MethodInfo(methodName); mt == nil {
		return remote.NewTransErrorWithMsg(remote.UnknownMethod, fmt.Sprintf("unknown method %s", methodName))
	}

	// unknown method doesn't set methodName for RPCInfo.To(), or lead inconsistent with old version
	rpcinfo.AsMutableEndpointInfo(ri.To()).SetMethod(methodName)
	return nil
}

// NewDataIfNeeded
func NewDataIfNeeded(methodName string, message remote.Message) error {
	// ri := message.RPCInfo()
	// ink := ri.Invocation()
	// TODO: 没懂
	return nil
}
