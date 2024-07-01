package configs

type Server struct {
	LAddr              string `yaml:"laddr"`
	VAddr              string `yaml:"vaddr"`
	Kleepalive         int    `yaml:"kleepalive"`
	JwtSecret          string `yaml:"jwt_secret"`
	JwtTimeout         int    `yaml:"jwt_timeout"`
	BattleSpaceDieTime int64  `yaml:"battle_space_die_time"`
	LogsPath           string `yaml:"logs_path"`
}
