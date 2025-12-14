package app

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitZapLogger() *zap.Logger {

	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.TimeKey = "timestamp"
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.StacktraceKey = ""

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zapConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
