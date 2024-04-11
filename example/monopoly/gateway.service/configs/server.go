package configs

type Server struct {
	LAddr             string `yaml:"laddr"`              // 监听地址
	LAddrServant      string `yaml:"laddr_servant"` 	 //监听内部服务
	VAddr             string `yaml:"vaddr"`              // 本服务虚拟地址
	EncryptionEnabled bool   `yaml:"encryption-enabled"` // 是否开启通信加密
	OnlineOfNumber    int    `yaml:"online-number"`      // 在线人数限制
	RequestTimeoutMax int32  `yaml:"request-timeout-max"`
}
