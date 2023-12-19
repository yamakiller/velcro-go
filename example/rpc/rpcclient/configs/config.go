package configs

import "github.com/yamakiller/velcro-go/envs"

type Config struct {
	TargetAddr              string `yaml:"target-addr"`
	ClientConnectionTimeout int    `yaml:"client-connection-timeout"`
	ClientNumber            int    `yaml:"client-number"`
	IntervalSecond          int    `yaml:"interval-second"`
	ScreenRefreshFrequency  int    `yaml:"screen-refresh-frequency"`
	PostData                string `yaml:"post-data"`
}

func (c *Config) IEnv() envs.IEnv {
	return &envs.YAMLEnv{}
}