package codec

import (
	"errors"
	"fmt"

	"github.com/yamakiller/velcro-go/rpc/utils/rpcinfo"
)

const (
	// FrontMask is used in protocol sniffing.
	FrontMask = 0x0000ffff
)

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
		ink.SetPackageName(svcInfo.GetPackageName())
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
