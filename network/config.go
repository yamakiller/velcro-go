package network

const (
	localAddress = "nonhost"
)

type Config struct {
	Producer   ProducerWidthClientSystem
	Kleepalive int32  //网络超时(单位毫秒)
	VAddr      string // 虚地址标记
}

func defaultConfig() *Config {
	return &Config{
		Kleepalive: 2000,
		VAddr:      localAddress,
	}
}
