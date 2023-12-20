package files

import "os"

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
