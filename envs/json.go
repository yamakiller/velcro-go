package envs

import (
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/files"
)

// JSONEnv doc
// @Summary  json fromat Environment variable loading interface
// @Struct JSONEnv
type JSONEnv struct {
	en map[string]interface{}
}

// Initial doc
// @Method Initial @Summary Initial json environment variable manager
func (slf *JSONEnv) Initial() {
	slf.en = make(map[string]interface{})
}

// Load doc
// @Summary file load to json env
// @Method Load
// @Param  (string) map key
// @Param  (string) file name
// @Param  (interface{}) env object
// @Return (error) load fail return error
func (slf *JSONEnv) Load(key string, fileName string, out interface{}) error {
	d, e := files.Instance().GetDataFromFile(fileName)
	if e != nil {
		goto fail
	}

	e = utils.JSONUnSerialize(d, out)
	if e != nil {
		goto fail
	}

	slf.en[key] = out
	return nil
fail:
	slf.en[key] = out
	return e
}

// UnLoad doc
// @Summary
// @Method UnLoad
func (slf *JSONEnv) UnLoad() {
	slf.en = nil
}

// Remove doc
// @Summary remove environment variable object
// @Method Remove
// @Param  (string) map key
func (slf *JSONEnv) Remove(key string) {
	if _, ok := slf.en[key]; ok {
		delete(slf.en, key)
	}
}

// Get doc
// @Summary Returns environment variable object
// @Method Get
// @Param  (string) map key
// @Return (interface{}) environment variable object
func (slf *JSONEnv) Get(key string) interface{} {
	v, ok := slf.en[key]
	if !ok {
		return nil
	}
	return v
}
