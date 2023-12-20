package envs

import (
	"encoding/xml"

	"github.com/yamakiller/velcro-go/utils/files"
)

// XMLEnv doc
// @Summary xml fromat Environment variable loading interface
// @Struct XMLEnv
type XMLEnv struct {
	en map[string]interface{}
}

// Initial doc
// @Method Initial @Summary Initial xml environment variable manager
func (slf *XMLEnv) Initial() {
	slf.en = make(map[string]interface{})
}

// Load dock
// @Summary file load to xml env
// @Method Load
// @Param  (string) map key
// @Param  (string) file name
// @Param  (interface{}) env object
// @Return (error) load fail return error
func (slf *XMLEnv) Load(key string, fileName string, out interface{}) error {
	d, e := files.Instance().GetDataFromFile(fileName)
	if e != nil {
		goto fail
	}

	e = xml.Unmarshal(d, out)
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
func (slf *XMLEnv) UnLoad() {
	slf.en = nil
}

// Remove doc
// @Summary remove environment variable object
// @Method Remove
// @Param  (string) map key
func (slf *XMLEnv) Remove(key string) {
	if _, ok := slf.en[key]; ok {
		delete(slf.en, key)
	}
}

// Get doc
// @Summary Returns environment variable object
// @Method Get
// @Param  (string) map key
// @Return (interface{}) environment variable object
func (slf *XMLEnv) Get(key string) interface{} {
	v, ok := slf.en[key]
	if !ok {
		return nil
	}
	return v
}
