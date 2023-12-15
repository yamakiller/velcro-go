package gateway

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/yamakiller/velcro-go/logs"
)

func produceLogger() logs.LogAgent {
	logLevel := logrus.DebugLevel
	if os.Getenv("DEBUG") != "" {
		logLevel = logrus.InfoLevel
	}

	logDir := getLogDir()
	pLogHandle := logs.SpawnFileLogrus(logLevel, logDir, "gateway")

	//丢弃屏幕输出
	pLogHandle.SetOutput(io.Discard)
	logAgent := &logs.DefaultAgent{}
	logAgent.WithHandle(pLogHandle)

	return logAgent
}

func getLogDir() string {
	ex, _ := os.Executable()

	exPath := filepath.Dir(ex)
	logDir := filepath.Join(exPath, "monitor/logs")

	if !isDirExits(logDir) {
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return ""
		}
	}

	return logDir
}

func isDirExits(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
