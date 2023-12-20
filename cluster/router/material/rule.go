package material

// Rule 角色
type Rule struct {
	RuleID   int32  `yaml:"rule-id"`   // 角色ID,最好适应掩码
	RuleName string `yaml:"rule-name"` // 角色名称
	RuleDesc string `yaml:"rule-desc"` // 角色说明
}
