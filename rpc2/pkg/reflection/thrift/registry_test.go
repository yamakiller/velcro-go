package thriftreflection

import (
	"testing"

	"github.com/cloudwego/thriftgo/reflection"
	"github.com/yamakiller/velcro-go/utils/test"
)

func TestRegistry(t *testing.T) {
	GlobalFiles.Register(&reflection.FileDescriptor{
		Filename: "testa",
	})
	test.Assert(t, GlobalFiles.GetFileDescriptors()["testa"] != nil)

	f := NewFiles()
	f.Register(&reflection.FileDescriptor{
		Filename: "testb",
	})
	test.Assert(t, f.GetFileDescriptors()["testb"] != nil)
}
