package registers

import (
	"fmt"
	"reflect"
)

var (
	rsm = make(map[string]reflect.Type)
)

func Register(name string, c interface{}) {
	rsm[name] = reflect.TypeOf(c).Elem()
}

func New(name string) (interface{}, error) {
	var c interface{}
	var err error
	if v, ok := rsm[name]; ok {
		c = reflect.New(v).Interface()
		return c, nil
	} else {
		err = fmt.Errorf("not found %s struct", name)
	}
	return nil, err
}

func CheckElem(name string) bool {
	if _, ok := rsm[name]; ok {
		return true
	}
	return false
}
