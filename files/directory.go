package files

import (
	"os"
	"runtime"
	"strings"
)

//Directory doc
//@Summary Virtual Directory
//@Struct Directory
//@Member (string) Base Directory
//@Member
type Directory struct {
	rootPath string
	wildCard string
}

//Initial doc
//@Summary initialization Directory
//@Method Initial
func (slf *Directory) Initial() {

	if runtime.GOOS == "windows" {
		slf.wildCard = "\\"
	} else {
		slf.wildCard = "/"
	}

	currPath, _ := os.Getwd()
	slf.WithRoot(currPath)
}

//WithRoot doc
//@Summary Setting Root path
//@Method WithRoot
//@Param  (string) path
func (slf *Directory) WithRoot(path string) {
	slf.rootPath = path
	if strings.HasSuffix(slf.rootPath, slf.wildCard) {
		slf.rootPath = slf.rootPath[:len(slf.rootPath)-1]
	}
}

//GetFullPathName doc
//@Summary Return Full path and file name
//@Method GetFullPathName
//@Return (string) Full path and file name
func (slf *Directory) GetFullPathName(filePath string) string {
	if strings.HasPrefix(filePath, slf.rootPath) {
		return filePath
	}

	return slf.rootPath + slf.wildCard + filePath
}
