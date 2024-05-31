package configs

type Server struct {
	LAddr      string `yaml:"laddr"`
	VAddr      string `yaml:"vaddr"`
	Kleepalive int    `yaml:"kleepalive"`
	JwtSecret  string `yaml:"jwt_secret"`
	JwtTimeout int    `yaml:"jwt_timeout"`
}
