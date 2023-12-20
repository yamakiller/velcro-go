package material

type Router struct {
	VAddress  string   `yaml:"virtual-address"`
	Rules     []int32  `yaml:"rules"`
	Commands  []string `yaml:"commands"`
	Endpoints []string `yaml:"endpoints"`
}
