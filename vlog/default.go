package vlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/yamakiller/velcro-go/utils/files"
)

var logger FullLogger = &defaultLogger{
	level:  LevelTrace,
	stdlog: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
}

func SetLogFile(logPath string, logName string) error {

	if logPath == "" {
		logPath, _ = os.Executable()
		logPath = filepath.Dir(logPath)
		logPath = path.Join(logPath, "logs")
	}

	if logName == "" {
		logName = "log"
	}
	baseLogPath := path.Join(logPath, logName)
	if !files.IsDirExits(logPath) {
		if err := files.MkdirAll(logPath); err != nil {
			return err
		}
	}

	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)

	if err != nil {
		return err
	}
	mw := io.MultiWriter(os.Stderr, writer)
	SetOutput(mw)
	return nil
}

// SetOutput sets the output of default logger. By default, it is stderr.
func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

// SetLevel sets the level of logs below which logs will not be output.
// The default log level is LevelTrace.
// Note that this method is not concurrent-safe.
func SetLevel(lv Level) {
	logger.SetLevel(lv)
}

// DefaultLogger return the default logger for velcro.
func DefaultLogger() FullLogger {
	return logger
}

// SetLogger sets the default logger.
// Note that this method is not concurrent-safe and must not be called
// after the use of DefaultLogger and global functions in this package.
func SetLogger(v FullLogger) {
	logger = v
}

// Fatal calls the default logger's Fatal method and then os.Exit(1).
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Error calls the default logger's Error method.
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Warn calls the default logger's Warn method.
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

// Notice calls the default logger's Notice method.
func Notice(v ...interface{}) {
	logger.Notice(v...)
}

// Info calls the default logger's Info method.
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Debug calls the default logger's Debug method.
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Trace calls the default logger's Trace method.
func Trace(v ...interface{}) {
	logger.Trace(v...)
}

// Fatalf calls the default logger's Fatalf method and then os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// Errorf calls the default logger's Errorf method.
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Warnf calls the default logger's Warnf method.
func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

// Noticef calls the default logger's Noticef method.
func Noticef(format string, v ...interface{}) {
	logger.Noticef(format, v...)
}

// Infof calls the default logger's Infof method.
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Debugf calls the default logger's Debugf method.
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Tracef calls the default logger's Tracef method.
func Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}

// ContextFatalf calls the default logger's ContextFatalf method and then os.Exit(1).
func ContextFatalf(ctx context.Context, format string, v ...interface{}) {
	logger.ContextFatalf(ctx, format, v...)
}

// ContextErrorf calls the default logger's ContextErrorf method.
func ContextErrorf(ctx context.Context, format string, v ...interface{}) {
	logger.ContextErrorf(ctx, format, v...)
}

// ContextWarnf calls the default logger's ContextWarnf method.
func ContextWarnf(ctx context.Context, format string, v ...interface{}) {
	logger.ContextWarnf(ctx, format, v...)
}

// ContextNoticef calls the default logger's ContextNoticef method.
func ContextNoticef(ctx context.Context, format string, v ...interface{}) {
	logger.ContextNoticef(ctx, format, v...)
}

// ContextInfof calls the default logger's ContextInfof method.
func ContextInfof(ctx context.Context, format string, v ...interface{}) {
	logger.ContextInfof(ctx, format, v...)
}

// ContextDebugf calls the default logger's ContextDebugf method.
func ContextDebugf(ctx context.Context, format string, v ...interface{}) {
	logger.ContextDebugf(ctx, format, v...)
}

// ContextTracef calls the default logger's ContextTracef method.
func ContextTracef(ctx context.Context, format string, v ...interface{}) {
	logger.ContextTracef(ctx, format, v...)
}

type defaultLogger struct {
	stdlog *log.Logger
	level  Level
}

func (ll *defaultLogger) SetOutput(w io.Writer) {
	ll.stdlog.SetOutput(w)
}

func (ll *defaultLogger) SetLevel(lv Level) {
	ll.level = lv
}

func (ll *defaultLogger) logf(lv Level, format *string, v ...interface{}) {
	if ll.level > lv {
		return
	}
	msg := lv.toString()
	if format != nil {
		msg += fmt.Sprintf(*format, v...)
	} else {
		msg += fmt.Sprint(v...)
	}
	ll.stdlog.Output(4, msg)
	if lv == LevelFatal {
		os.Exit(1)
	}
}

func (ll *defaultLogger) Fatal(v ...interface{}) {
	ll.logf(LevelFatal, nil, v...)
}

func (ll *defaultLogger) Error(v ...interface{}) {
	ll.logf(LevelError, nil, v...)
}

func (ll *defaultLogger) Warn(v ...interface{}) {
	ll.logf(LevelWarn, nil, v...)
}

func (ll *defaultLogger) Notice(v ...interface{}) {
	ll.logf(LevelNotice, nil, v...)
}

func (ll *defaultLogger) Info(v ...interface{}) {
	ll.logf(LevelInfo, nil, v...)
}

func (ll *defaultLogger) Debug(v ...interface{}) {
	ll.logf(LevelDebug, nil, v...)
}

func (ll *defaultLogger) Trace(v ...interface{}) {
	ll.logf(LevelTrace, nil, v...)
}

func (ll *defaultLogger) Fatalf(format string, v ...interface{}) {
	ll.logf(LevelFatal, &format, v...)
}

func (ll *defaultLogger) Errorf(format string, v ...interface{}) {
	ll.logf(LevelError, &format, v...)
}

func (ll *defaultLogger) Warnf(format string, v ...interface{}) {
	ll.logf(LevelWarn, &format, v...)
}

func (ll *defaultLogger) Noticef(format string, v ...interface{}) {
	ll.logf(LevelNotice, &format, v...)
}

func (ll *defaultLogger) Infof(format string, v ...interface{}) {
	ll.logf(LevelInfo, &format, v...)
}

func (ll *defaultLogger) Debugf(format string, v ...interface{}) {
	ll.logf(LevelDebug, &format, v...)
}

func (ll *defaultLogger) Tracef(format string, v ...interface{}) {
	ll.logf(LevelTrace, &format, v...)
}

func (ll *defaultLogger) ContextFatalf(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelFatal, &format, v...)
}

func (ll *defaultLogger) ContextErrorf(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelError, &format, v...)
}

func (ll *defaultLogger) ContextWarnf(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelWarn, &format, v...)
}

func (ll *defaultLogger) ContextNoticef(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelNotice, &format, v...)
}

func (ll *defaultLogger) ContextInfof(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelInfo, &format, v...)
}

func (ll *defaultLogger) ContextDebugf(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelDebug, &format, v...)
}

func (ll *defaultLogger) ContextTracef(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelTrace, &format, v...)
}
