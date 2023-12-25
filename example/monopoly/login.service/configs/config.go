package configs

import (
	"os"
	"path/filepath"
)

func GetLocalConfigFilePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	exPath := filepath.Dir(ex)
	cfgFilePath := filepath.Join(exPath, "config.yml")

	return cfgFilePath, nil
}

func GetLogDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)
	logDir := filepath.Join(exPath, "logs")

	return logDir, nil
}