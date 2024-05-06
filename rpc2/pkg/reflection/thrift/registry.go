package thriftreflection

import (
	"sync"

	"github.com/cloudwego/thriftgo/reflection"
)

// Files 是一个用于查找或迭代文件及其中包含的描述符的注册表.
type Files struct {
	filesByPath map[string]*reflection.FileDescriptor
}

func NewFiles() *Files {
	return &Files{filesByPath: map[string]*reflection.FileDescriptor{}}
}

var globalMutex sync.RWMutex

// GlobalFiles 是文件描述符的全局注册表.
var GlobalFiles = NewFiles()

// RegisterIDL 为生成的代码提供将其反射信息注册到 GlobalFiles 的功能.
func RegisterIDL(bytes []byte) {
	desc := reflection.Decode(bytes)
	GlobalFiles.Register(desc)
}

// Register 将输入 FileDescriptor 注册到 *Files 类型变量.
func (f *Files) Register(desc *reflection.FileDescriptor) {
	if f == GlobalFiles {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	// TODO: check conflict
	f.filesByPath[desc.Filename] = desc
}

// GetFileDescriptors 返回内部注册的反射 FileDescriptors.
func (f *Files) GetFileDescriptors() map[string]*reflection.FileDescriptor {
	if f == GlobalFiles {
		globalMutex.RLock()
		defer globalMutex.RUnlock()
		m := make(map[string]*reflection.FileDescriptor, len(f.filesByPath))
		for k, v := range f.filesByPath {
			m[k] = v
		}
		return m
	}
	return f.filesByPath
}
