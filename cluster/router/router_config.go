package router

type RouterConfig struct {
	URI              string `yaml:"uri"`
	ProxyDialTimeout int32  `yaml:"dialtimeout"`
	ProxyKleepalive  int32  `yaml:"kleepalive"`
	ProxyAlgorithm   string `yaml:"algorithm"`
}
