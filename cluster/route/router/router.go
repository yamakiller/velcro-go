package router

type Router struct {
	RuleList  []Rule       `yaml:"rules"`
	RouteList []RouteGroup `yaml:"routes"`
}
