package configs

type Redis struct {
	Addr    string       `yaml:"address"`
	Pwd     string       `yaml:"password"`
	Timeout RedisTimeout `yaml:"timeout"`
}

type RedisTimeout struct {
	Dial  int `yaml:"dial"`
	Read  int `yaml:"read"`
	Write int `yaml:"write"`
}
