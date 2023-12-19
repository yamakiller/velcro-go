package router

type RouteGroup struct {
	Label     string   `yaml:"label"`
	VxAddress string   `yaml:"address"`
	Rules     []int32  `yaml:"rules"`
	Commands  []string `yaml:"commands"`
	Endpoints []string `yaml:"endpoints"`
}
