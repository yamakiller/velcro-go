package configs

import (
	"github.com/yamakiller/velcro-go/cluster/router"
)

type Config struct {
	Server        Server              `yaml:"server"`
	Router        router.RouterConfig `yaml:"router"`          // 路由文件地址
	LogRemoteAddr string              `yaml:"log_remote_addr"` // 日志远程地址
}
