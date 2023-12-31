package router

type RouterConfig struct {
	URI         string `yaml:"uri"`
	DialTimeout int32  `yaml:"dialtimeout"`
	Kleepalive  int32  `yaml:"kleepalive"`
	Algorithm   string `yaml:"algorithm"`
}
