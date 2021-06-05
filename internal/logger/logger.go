package logger

import (
	"errors"
	"go.uber.org/zap/zapcore"
	"os"

	"go.uber.org/zap"
)

var (
	// TODO: User viper here
	LogMode        = "dev"
	LogPath        = "log/router.log"
	LogConfig      zap.Config
	LogAtomicLevel zap.AtomicLevel
	LogLevel       zapcore.Level
	Logger         *zap.Logger
)

func init() {
	err := os.Remove(LogPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return
		}
	}

	LogConfig = zap.NewDevelopmentConfig()
	if LogMode == "prod" || LogMode == "production" {
		LogConfig = zap.NewProductionConfig()
	}

	// TODO: this should be configurable
	LogLevel = zapcore.WarnLevel
	LogAtomicLevel = zap.NewAtomicLevel()
	LogAtomicLevel.SetLevel(LogLevel)

	LogConfig.OutputPaths = []string{
		LogPath,
	}
	LogConfig.Level = LogAtomicLevel
}

func InitLogger() {
	Logger, _ = LogConfig.Build()
}

func GetLogger() *zap.Logger {
	if Logger == nil {
		InitLogger()
	}
	return Logger
}
