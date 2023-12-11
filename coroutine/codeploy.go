package coroutine

import "math"

const (
	//EnvKey @Summary coroutine pool env name
	EnvKey = "coroutine pool"
)

//Deploy doc
//@Summary coroutine pool deploy informat (json)
//@Struct Deploy
type Deploy struct {
	Max  int `xml:"max" yaml:"max" json:"max"`
	Min  int `xml:"min" yaml:"min" json:"min"`
	Task int `xml:"task-limit" yaml:"task-limit" json:"task-limit"`
}

//NewDefault doc
//@Summary create default coroutine pool deploy informat
//@Method NewDefault
func NewDefault() *Deploy {
	return &Deploy{math.MaxInt16, 32, 64}
}
