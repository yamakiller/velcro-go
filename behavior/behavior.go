package behavior

const (
	COMPOSITE = "composite"
	DECORATOR = "decorator"
	ACTION    = "action"
	CONDITION = "condition"
)

// Returning status
type Status uint8

const (
	SUCCESS Status = 1
	FAILURE Status = 2
	RUNNING Status = 3
	ERROR   Status = 4
)

func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
