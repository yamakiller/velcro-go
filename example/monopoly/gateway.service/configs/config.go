package configs

import "github.com/yamakiller/velcro-go/cluster/router"

type Config struct {
	Server Server              `yaml:"server"`
	Router router.RouterConfig `yaml:"router"` // 路由文件地址
}
