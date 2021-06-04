package logger

import (
	"go.uber.org/zap"
)

var (
	Logger *zap.Logger
)

func InitLogger() {
	cfg := zap.NewProductionConfig()
	// TODO: this should be configurable
	cfg.OutputPaths = []string{
		"log/router.log",
	}

	Logger, _ = cfg.Build()
}

func GetLogger() *zap.Logger {
	if Logger == nil {
		InitLogger()
	}
	return Logger
}
