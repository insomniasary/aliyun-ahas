package logger

import (
	"log"
	"os"
	"strings"

	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	AhasLogFile = "ahas.log"
)

var (
	ahasLogger *zap.Logger
)

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer l.Sync()
	ahasLogger = l
}

func toZapLevel(level logging.Level) zapcore.Level {
	switch level {
	case logging.ErrorLevel:
		return zapcore.ErrorLevel
	case logging.InfoLevel:
		return zapcore.InfoLevel
	case logging.DebugLevel:
		return zapcore.DebugLevel
	case logging.WarnLevel:
		return zapcore.WarnLevel
	case logging.PanicLevel:
		return zapcore.PanicLevel
	case logging.FatalLevel:
		return zapcore.FatalLevel
	default:
	}
	return zapcore.InfoLevel
}

func InitLoggerDefault() error {
	logDir := config.LogBaseDir()
	if logDir == "" {
		return nil
	}
	logDir = addSeparatorIfNeeded(logDir)
	path := logDir + AhasLogFile

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    20, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
	})
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		toZapLevel(logging.GetGlobalLoggerLevel()),
	)
	logger := zap.New(core)
	ahasLogger = logger

	return nil
}

func addSeparatorIfNeeded(path string) string {
	s := string(os.PathSeparator)
	if !strings.HasSuffix(path, s) {
		return path + s
	}
	return path
}

func Debug(v ...interface{}) {
	if logging.DebugLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	if logging.DebugLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Debugf(format, v...)
}

func Info(v ...interface{}) {
	if logging.InfoLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Info(v...)
}

func Infof(format string, v ...interface{}) {
	if logging.InfoLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Infof(format, v...)
}

func Warn(v ...interface{}) {
	if logging.WarnLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	if logging.WarnLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Warnf(format, v...)
}

func Error(v ...interface{}) {
	if logging.ErrorLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Error(v...)
}

func Errorf(format string, v ...interface{}) {
	if logging.ErrorLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	if logging.FatalLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	if logging.FatalLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Fatalf(format, v...)
}

func Panic(v ...interface{}) {
	if logging.PanicLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Panic(v...)
}

func Panicf(format string, v ...interface{}) {
	if logging.PanicLevel < logging.GetGlobalLoggerLevel() || ahasLogger == nil {
		return
	}
	ahasLogger.Sugar().Panicf(format, v...)
}
