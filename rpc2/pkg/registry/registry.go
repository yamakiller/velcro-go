package registry

import (
	"net"
	"time"
)

type Registry interface {
	Register(info *Info) error
	Unregister(info *Info) error
}

// Info 用于注册.
// 这些字段只是建议的，具体使用取决于设计
type Info struct {
	// ServiceName will be set in velcro by default
	ServiceName string
	// Addr 默认情况下将在 velcro 中设置
	Addr net.Addr
	// PayloadCodec 默认情况下会在 velcro 中设置，如 thrift、protobuf.
	PayloadCodec string

	Weight    int
	StartTime time.Time
	WarmUp    time.Duration

	// extend other infos with Tags
	Tags map[string]string
}

var NoopRegistry = &noopRegistry{}

type noopRegistry struct{}

func (e noopRegistry) Register(*Info) error {
	return nil
}

func (e noopRegistry) Unregister(*Info) error {
	return nil
}
