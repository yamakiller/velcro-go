package configs

type Redis struct {
	Addr         string `yaml:"addr"`
	Pwd          string `yaml:"pwd"`
	DialTimeout  int    `yaml:"dial-timeout"`
	ReadTimeout  int    `yaml:"read-timeout"`
	WriteTimeout int    `yaml:"write-timeout"`
}