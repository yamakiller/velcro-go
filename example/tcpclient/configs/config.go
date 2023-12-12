package configs

type Config struct {
	TargetAddr              string `yaml:"target-addr"`
	ClientConnectionTimeout int    `yaml:"client-connection-timeout"`
	ClientNumber            int    `yaml:"client-number"`
	IntervalSecond          int    `yaml:"interval-second"`
	ScreenRefreshFrequency  int    `yaml:"screen-refresh-frequency"`
	PostData                string `yaml:"post-data"`
}
