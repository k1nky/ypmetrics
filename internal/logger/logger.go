package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zapLogger *zap.Logger
	level     zap.AtomicLevel
}

func New() *Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	config := zap.Config{
		Level:             level,
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
		level:     level,
	}
}

func (l *Logger) SetLevel(level string) error {
	levelValue, err := zapcore.ParseLevel(level)
	if err != nil {
		return err
	}
	l.level.SetLevel(levelValue)
	return nil
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.zapLogger.Debug(fmt.Sprintf(template, args...))
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(template, args...))
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.zapLogger.Error(fmt.Sprintf(template, args...))
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.zapLogger.Warn(fmt.Sprintf(template, args...))
}
