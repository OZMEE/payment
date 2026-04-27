package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() (*zap.Logger, error) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.Lock(os.Stderr),
			zap.NewAtomicLevelAt(zap.InfoLevel),
		),
	)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.ErrorOutput(zapcore.Lock(os.Stderr)))

	return logger, nil
}
