package hashring

import (
	"fmt"
	"hash"
)

//HashSum allows to use a builder pattern to create different HashFunc objects
type HashSum struct {
	_functions []func([]byte) []byte
}

//Use ...
func (slf *HashSum) Use(hashKeyFunc func(bytes []byte) (HashKey, error)) (HashFunc, error) {

	// build final hash function
	composed := func(bytes []byte) []byte {
		for _, f := range slf._functions {
			bytes = f(bytes)
		}
		return bytes
	}

	// check function composition for errors
	testResult := composed([]byte("test"))
	_, err := hashKeyFunc(testResult)
	if err != nil {
		const msg = "can't use given hash.Hash with given hashKeyFunc"
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	// build HashFunc
	return func(key []byte) HashKey {
		bytes := composed(key)
		hashKey, err := hashKeyFunc(bytes)
		if err != nil {
			// panic because we already checked HashSum earlier
			panic(fmt.Sprintf("hashKeyFunc failure: %v", err))
		}
		return hashKey
	}, nil
}

//FirstBytes 开始第n个字节
func (slf *HashSum) FirstBytes(n int) *HashSum {
	slf._functions = append(slf._functions, func(bytes []byte) []byte {
		return bytes[:n]
	})
	return slf
}

//LastBytes 最后第n个字节
func (slf *HashSum) LastBytes(n int) *HashSum {
	slf._functions = append(slf._functions, func(bytes []byte) []byte {
		return bytes[len(bytes)-n:]
	})
	return slf
}

//NewHash 创建一个 hash 对象
func NewHash(hasher func() hash.Hash) *HashSum {
	return &HashSum{
		_functions: []func(key []byte) []byte{
			func(key []byte) []byte {
				hash := hasher()
				hash.Write(key)
				return hash.Sum(nil)
			},
		},
	}
}
