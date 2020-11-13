package gateway

import (
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
)

func logDebug(args ...interface{}) {
	logger.Debug(args...)
}

func logWarn(args ...interface{}) {
	logger.Warn(args...)
}

func logInfo(args ...interface{}) {
	logger.Info(args...)
}

func logError(args ...interface{}) {
	logger.Error(args...)
}

func logDebugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func logInfof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func logWarnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func logErrorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}
