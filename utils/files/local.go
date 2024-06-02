package files

import (
	"os"
	"path/filepath"
)

func IsFileExist(fullPath string) bool {

	_, err := os.Stat(fullPath)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

func IsDirExits(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func NewLocalPathFull(fileName string) string {
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)
	return filepath.Join(exPath, fileName)
}

func MkdirAll(dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}
