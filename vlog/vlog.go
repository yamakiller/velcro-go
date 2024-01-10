package vlog

import (
	"context"
	"fmt"
	"io"
)

// FormatLogger is a logger interface that output logs with a format.
type FormatLogger interface {
	Fatalf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Noticef(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Tracef(format string, v ...interface{})
}

// Logger is a logger interface that provides logging function with levels.
type Logger interface {
	Trace(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Notice(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
}

// ContextLogger is a logger interface that accepts a context argument and output
// logs with a format.
type ContextLogger interface {
	ContextTracef(ctx context.Context, format string, v ...interface{})
	ContextDebugf(ctx context.Context, format string, v ...interface{})
	ContextInfof(ctx context.Context, format string, v ...interface{})
	ContextNoticef(ctx context.Context, format string, v ...interface{})
	ContextWarnf(ctx context.Context, format string, v ...interface{})
	ContextErrorf(ctx context.Context, format string, v ...interface{})
	ContextFatalf(ctx context.Context, format string, v ...interface{})
}

// Control provides methods to config a logger.
type Control interface {
	SetLevel(Level)
	SetOutput(io.Writer)
	SetElasticProducerPostmessage(string,func(string,[]byte)error)
}

// FullLogger is the combination of Logger, FormatLogger, ContextLogger and Control.
type FullLogger interface {
	Logger
	FormatLogger
	ContextLogger
	Control
}

// Level defines the priority of a log message.
// When a logger is configured with a level, any log message with a lower
// log level (smaller by integer comparison) will not be output.
type Level int

// The levels of logs.
const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelNotice
	LevelWarn
	LevelError
	LevelFatal
)

var strs = []string{
	"[Trace] ",
	"[Debug] ",
	"[Info] ",
	"[Notice] ",
	"[Warn] ",
	"[Error] ",
	"[Fatal] ",
}

func (lv Level) toString() string {
	if lv >= LevelTrace && lv <= LevelFatal {
		return strs[lv]
	}
	return fmt.Sprintf("[?%d] ", lv)
}
