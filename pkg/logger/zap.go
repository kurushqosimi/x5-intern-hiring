package logger

import (
	"go.uber.org/zap"
)

var globalLogger *zap.Logger

func New() *zap.Logger {
	config := zap.NewProductionConfig()
	config.DisableCaller = true
	config.DisableStacktrace = true
	logger, _ := config.Build()

	globalLogger = logger
	return logger
}

func Get() *zap.Logger {
	if globalLogger != nil {
		return globalLogger
	}

	return New()
}
