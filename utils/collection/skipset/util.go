package skipset

import (
	"github.com/yamakiller/velcro-go/utils/internal/wyhash"
	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
)

const (
	maxLevel            = 16
	p                   = 0.25
	defaultHighestLevel = 3
)

func hash(s string) uint64 {
	return wyhash.Sum64String(s)
}

//go:linkname cmpstring runtime.cmpstring
func cmpstring(a, b string) int

func randomLevel() int {
	level := 1
	for fastrand.Uint32n(1/p) == 0 {
		level++
	}
	if level > maxLevel {
		return maxLevel
	}
	return level
}
