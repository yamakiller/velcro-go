package server

import (
	"net"
	"time"

	"github.com/yamakiller/velcro-go/utils"
)

const (
	defaultExitWaitTime          = 5 * time.Second       // 默认的推出等待时间
	defaultAcceptFailedDelayTime = 10 * time.Millisecond // 默认Accept 失败延迟关闭时间
	defaultConnectionIdleTime    = 10 * time.Minute      // 默认连接空闲时间
)

// 默认服务地址
var defaultAddress = utils.NewNetAddr("tcp", ":8888")

type Config struct {
	Address net.Addr

	// 服务器等待允许正常关闭任何现有连接的持续时间.
	ExitWaitTime time.Duration

	// 接受连接发生错误后服务器等待的时间.
	AcceptFailedDelayTime time.Duration

	// 接受的连接等待读取或写入数据的时间，仅在NIO下有效.
	MaxConnectionIdleTime time.Duration
}

// NewConfig 创建一个默认的配置文件
func NewConfig() *Config {
	return &Config{
		Address:               defaultAddress,
		ExitWaitTime:          defaultExitWaitTime,
		AcceptFailedDelayTime: defaultAcceptFailedDelayTime,
		MaxConnectionIdleTime: defaultConnectionIdleTime,
	}
}
