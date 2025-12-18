package logger

import (
	"go.uber.org/zap"
)

func New() *zap.Logger {
	config := zap.NewProductionConfig()
	config.DisableCaller = true
	config.DisableStacktrace = true
	logger, _ := config.Build()

	return logger
}
