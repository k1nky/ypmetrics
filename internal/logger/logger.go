package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zapLogger *zap.Logger
}

func New() *Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encoderConfig,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}
	return &Logger{
		zapLogger: zap.Must(config.Build()),
	}
}

func (l *Logger) Debug(template string, args ...interface{}) {
	l.zapLogger.Debug(fmt.Sprintf(template, args...))
}

func (l *Logger) Info(template string, args ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(template, args...))
}

func (l *Logger) Error(template string, args ...interface{}) {
	l.zapLogger.Error(fmt.Sprintf(template, args...))
}
