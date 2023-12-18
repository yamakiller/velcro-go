package router

// 默认角色ID
const (
	// NONE_RULE_ID 初始角色
	NONE_RULE_ID = -2
	// 密钥完成角色
	KEYED_RULE_ID = -1
)

type Rule struct {
	RuleID   int32  `yaml:"rule-id"`   // 角色ID
	RuleName string `yaml:"rule-name"` // 角色名字
	RuleDesc string `yaml:"rule-desc"` // 角色说明
}

// RouteGroup 路由组
type RouteGroup struct {
	RouteId     string   `yaml:"route-id"`           // 路由ID
	RouteName   string   `yaml:"route-name"`         // 路由名称
	RouteDesc   string   `yaml:"route-desc"`         // 路由说明
	RuleList    []string `yaml:"route-rule-list"`    // 路由角色表
	TargetList  []Target `yaml:"route-target-list"`  // 路由目标表
	CommandList []string `yaml:"route-command-list"` // 路由指令表
}

type Target struct {
	TargetId      string `yaml:"target-id"`      // 目标ID
	TargetWeight  int64  `yaml:"target-weight"`  // 目标权重
	TargetAddress string `yaml:"target-address"` // 目标地址
}
