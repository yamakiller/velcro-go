package configs

type Server struct {
	LAddr      string `yaml:"laddr"`
	VAddr      string `yaml:"vaddr"`
	Kleepalive int    `yaml:"kleepalive"`
}
