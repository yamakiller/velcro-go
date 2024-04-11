package protocol

import (
	"reflect"
	"strings"

	"github.com/apache/thrift/lib/go/thrift"
)

func MessageName(message thrift.TStruct) string {
	return strings.Replace(reflect.TypeOf(message).String(), "*", "", 1)
}
