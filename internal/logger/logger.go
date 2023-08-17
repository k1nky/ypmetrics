package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type Logger struct {
	zapLogger *zap.Logger
}

func New() *Logger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
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

func (l *Logger) Info(template string, args ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(template, args...))
}

func (l *Logger) Error(template string, args ...interface{}) {
	l.zapLogger.Error(fmt.Sprintf(template, args...))
}
