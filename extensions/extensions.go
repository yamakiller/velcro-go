package extensions

import "sync/atomic"

type ExtensionID int32

var currentId int32

type Extension interface {
	ExtensionID() ExtensionID
}

type Extensions struct {
	_extensions []Extension
}

func NewExtensions() *Extensions {
	ex := &Extensions{
		_extensions: make([]Extension, 128),
	}

	return ex
}

func NextExtensionID() ExtensionID {
	id := atomic.AddInt32(&currentId, 1)

	return ExtensionID(id)
}

func (ex *Extensions) Get(id ExtensionID) Extension {
	return ex._extensions[id]
}

func (ex *Extensions) Register(extension Extension) {
	id := extension.ExtensionID()
	ex._extensions[id] = extension
}
