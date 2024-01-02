package serve

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/utils/files"
)

// ProduceLogger 产生日志对象
func ProduceLogger(name string) logs.LogAgent {
	logLevel := logrus.DebugLevel
	if os.Getenv("DEBUG") != "" {
		logLevel = logrus.InfoLevel
	}

	logDir := getLogDir()
	pLogHandle := logs.SpawnFileLogrus(logLevel, logDir, name+"Service")

	//丢弃屏幕输出
	pLogHandle.SetOutput(io.Discard)
	logAgent := &logs.DefaultAgent{}
	logAgent.WithHandle(pLogHandle)

	return logAgent
}

func getLogDir() string {
	logDir := files.NewLocalPathFull("monitor/logs")
	if !files.IsDirExits(logDir) {
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return ""
		}
	}

	return logDir
}
