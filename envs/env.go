package envs

import "sync"

var (
	oneEnv     sync.Once
	defaultEnv IEnv
)

//With doc
//@Summary Setting Environment variable object
//@Method With
//@Param (IEnv) new Environment variable object
func With(h IEnv) {
	defaultEnv = h
}

//Instance doc
//@Summary Environment variable object instance
//@Method Instance
//@Return (IEnv) an Environment variable object
func Instance() IEnv {
	oneEnv.Do(func() {
		if defaultEnv == nil {
			defaultEnv = &JSONEnv{}
		}
		defaultEnv.Initial()
	})

	return defaultEnv
}

//IEnv doc
//@Summary Environment variable loading interface
//@Interface IEnv
//@Method Load @Summary file load to environment variable
type IEnv interface {
	Initial()
	Load(key string, fileName string, out interface{}) error
	UnLoad()
	Remove(key string)
	Get(key string) interface{}
}
