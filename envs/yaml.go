package envs

import (
	yaml "gopkg.in/yaml.v2"

	"github.com/yamakiller/velcro-go/files"
)

// YAMLEnv docS
// @Summary yaml fromat Environment variable loading interface
// @Struct XMLEnv
type YAMLEnv struct {
	en map[string]interface{}
}

// Initial doc
// @Summary Initial yaml environment variable manager
// @Method Initial
func (slf *YAMLEnv) Initial() {
	slf.en = make(map[string]interface{})
}

// Load doc
// @Summary file load to yaml env
// @Method Load
// @Param  (string) map key
// @Param  (string) file name
// @Param  (interface{}) env object
// @Return (error) load fail return error
func (slf *YAMLEnv) Load(key string, fileName string, out interface{}) error {
	d, e := files.Instance().GetDataFromFile(fileName)
	if e != nil {
		goto fail
	}

	e = yaml.Unmarshal(d, out)
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
func (slf *YAMLEnv) UnLoad() {
	slf.en = nil
}

// Remove doc
// @Summary remove environment variable object
// @Method Remove
// @Param  (string) map key
func (slf *YAMLEnv) Remove(key string) {
	if _, ok := slf.en[key]; ok {
		delete(slf.en, key)
	}
}

// Get doc
// @Summary Returns environment variable object
// @Method Get
// @Param  (string) map key
// @Return (interface{}) environment variable object
func (slf *YAMLEnv) Get(key string) interface{} {
	v, ok := slf.en[key]
	if !ok {
		return nil
	}
	return v
}
