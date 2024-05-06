package trans

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/rpc2/pkg/serviceinfo"
)

var defaultReadMoreTimeout = 5 * time.Millisecond

// ExtAddon trans苦战需要实现的接口.
// 通常, 如果我们想扩展传输层, 我们需要实现 trans_handler.go 中定义的 trans 接口.
// 事实上大多数代码逻辑在同一模式下都是相似的，因此ExtAddon接口是需要单独实现的差异化部分.
// 默认的通用反式实现在default_client_handler.go和default_server_handler.go
type ExtAddon interface {
	SetReadTimeout(ctx context.Context, c net.Conn, cfg rpcinfo.RPCConfig, role remote.RPCRole)
	NewWriteByteBuffer(ctx context.Context, c net.Conn, msg remote.Message) remote.ByteBuffer
	NewReadByteBuffer(ctx context.Context, c net.Conn, msg remote.Message) remote.ByteBuffer
	ReleaseBuffer(remote.ByteBuffer, error) error
	IsTimeoutErr(error) bool
	IsRemoteClosedErr(error) bool
}

// GetReadTimeout 对超时进行修改正, 考虑到中间层的延时
func GetReadTimeout(cfg rpcinfo.RPCConfig) time.Duration {
	if cfg.RPCTimeout() <= 0 {
		return 0
	}

	return cfg.RPCTimeout() + defaultReadMoreTimeout
}

// GetMethodInfo 用于通过方法名从serviceinfo.MethodInfo获取方法信息.
func GetMethodInfo(ri rpcinfo.RPCInfo, svcInfo *serviceinfo.ServiceInfo) (serviceinfo.MethodInfo, error) {
	methodName := ri.Invocation().MethodName()
	methodInfo := svcInfo.MethodInfo(methodName)
	if methodInfo != nil {
		return methodInfo, nil
	}
	return nil, remote.NewTransErrorWithMsg(remote.UnknownMethod, fmt.Sprintf("unknown method %s", methodName))
}

// MuxEnabledFlag 用于判断某个serverHandlerFactory是否是复用的
type MuxEnabledFlag interface {
	MuxEnabled() bool
}

// GetDefaultSvcInfo 用于从服务映射表中获取一个 ServiceInfo, 该地图应该有一个 ServiceInfo.
func GetDefaultSvcInfo(svcMap map[string]*serviceinfo.ServiceInfo) *serviceinfo.ServiceInfo {
	for _, svcInfo := range svcMap {
		return svcInfo
	}
	return nil
}
