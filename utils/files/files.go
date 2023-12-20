package files

import (
	"os"
	"sync"
	"sync/atomic"
)

// Files doc
// @Summary files system
// @Struct Files
type Files struct {
	_dir        Directory
	_cached     sync.Map
	_cachedFils int64
	_cacheMem   int64
}

// Initial doc
// @Summary Initial files system
// @Method Initial
func (slf *Files) Initial() {
	slf._dir.Initial()
}

// Close doc
// @Summary Close System and clear data
// @Method Close
func (slf *Files) Close() {

	var key []string
	slf._cached.Range(func(k, v interface{}) bool {
		key = append(key, k.(string))
		return true
	})

	for _, k := range key {
		slf._cached.Delete(k)
	}
	slf._cachedFils = 0
	slf._cacheMem = 0
	key = nil
}

// IsFileExist doc
// @Summary files is exist
// @Method IsFileExist
// @Param (string) file path and name
// @Return (bool) exist:true not exits: false
func (slf *Files) IsFileExist(fullPath string) bool {
	h := slf.getCache(fullPath)
	if h != nil {
		return true
	}

	_, err := os.Stat(fullPath)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

// DeleteFile doc
// @Summary delete file and remove cache
// @Method DeleteFile
// @Param  (string) full path and file name
// @retur  (error) delete fail error
func (slf *Files) DeleteFile(fullPath string) error {
	h := slf.getCache(fullPath)
	if h != nil {
		atomic.AddInt64(&slf._cacheMem, -int64(h._size))
		atomic.AddInt64(&slf._cachedFils, -1)
		slf._cached.Delete(fullPath)
	}
	err := os.Remove(fullPath)
	if err != nil {
		return err
	}
	return nil
}

// GetRoot doc
// @Summary Return root dir
// @Method GetRoot
// @Return (string) dir
func (slf *Files) GetRoot() string {
	return slf._dir.rootPath
}

// GetCacheFiles doc
// @Summary Return Cache file of number
// @Method GetCacheFiles
// @Return (int) file of number
func (slf *Files) GetCacheFiles() int {
	return int(atomic.LoadInt64(&slf._cachedFils))
}

// GetCacheMem doc
// @Summary Return Cache memory
// @Method GetCacheMem
// @Return (int64) Cache size
func (slf *Files) GetCacheMem() int64 {
	return atomic.LoadInt64(&slf._cacheMem)
}

// GetDataFromFile doc
// @Summary Retrun File data
// @Method GetDataFromFile
// @Param  (string) full path and file name
// @Return ([]byte) file data
// @Return (error)
func (slf *Files) GetDataFromFile(fullPath string) ([]byte, error) {

	d, e := os.ReadFile(fullPath)
	if e != nil {
		return nil, e
	}

	return d, nil
}

// GetDataFromCacheFile doc
// @Summary Use Cache or Read File Return data
// @Method GetDataFromCacheFile
// @Param  (string) full path and file name
// @Return (*FileHandle) file handle
// @Return (error)
func (slf *Files) GetDataFromCacheFile(fullPath string) (*FileHandle, error) {
	h := slf.getCache(fullPath)
	if h != nil {
		return h, nil
	}

	d, e := os.ReadFile(fullPath)
	if e != nil {
		return nil, e
	}

	h = &FileHandle{d, len(d)}
	oh, ok := slf._cached.LoadOrStore(fullPath, h)
	if ok {
		return oh.(*FileHandle), nil
	}
	atomic.AddInt64(&slf._cacheMem, int64(len(d)))
	atomic.AddInt64(&slf._cachedFils, 1)
	return h, nil
}

// WithRoot doc
// @Summarysetting system root dir
// @Method WithRoot
// @Param (string) path
func (slf *Files) WithRoot(path string) {
	slf._dir.WithRoot(path)
}

// GetFullPathForFilename doc
// @Summary Returns file full path and file name
// @Method GetFullPathForFilename
// @Param  (string) file path
// @Return (string) full path and file name
func (slf *Files) GetFullPathForFilename(filePath string) string {
	return slf._dir.GetFullPathName(filePath)
}

// getCache doc
// @Summary Return file cache
// @Method getCache
// @Return (*FileHandle) file cache handle
func (slf *Files) getCache(fullPath string) *FileHandle {
	r, ok := slf._cached.Load(fullPath)
	if !ok {
		return nil
	}

	h, ok := r.(*FileHandle)
	if !ok {
		return nil
	}
	return h
}

var (
	oneFiles    sync.Once
	defaultFile *Files
)

// Instance doc
// @Summary Files System instance
// @Method Instance
// @Return (*Files) Files System object
func Instance() *Files {
	oneFiles.Do(func() {
		defaultFile = &Files{}
		defaultFile.Initial()
	})

	return defaultFile
}
