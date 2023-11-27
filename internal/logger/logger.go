// Пакет logger представляет инстуременты для ведения журнала событйи приложения.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger логгер
type Logger struct {
	zapLogger *zap.Logger
	level     zap.AtomicLevel
}

// New возвращает новый логгер.
func New() *Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	config := zap.Config{
		Level:       level,
		Development: false,
		// не отображаем имя файла, в котором был вызван метод логгера
		// в нашем случае это всегда будет один и тот же файл.
		DisableCaller: true,
		// выводим stacktrace
		DisableStacktrace: false,
		Sampling:          nil,
		// формат лога "читаемые сообщения из консоли"
		// лучше использовать json
		Encoding:      "console",
		EncoderConfig: encoderConfig,
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

// SetLevel задает уровень логирования. Полный список допустимых уровней
// можно посмотреть в документации пакета go.uber.org/zap.
func (l *Logger) SetLevel(level string) error {
	levelValue, err := zapcore.ParseLevel(level)
	if err != nil {
		return err
	}
	l.level.SetLevel(levelValue)
	return nil
}

// Debugf записывает сообщение уровня debug.
// Сообщение формируется с помощью fmt.Sprintf(template, args...).
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.zapLogger.Debug(fmt.Sprintf(template, args...))
}

// Infof записывает сообщение уровня info.
// Сообщение формируется с помощью fmt.Sprintf(template, args...).
func (l *Logger) Infof(template string, args ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(template, args...))
}

// Warnf записывает сообщение уровня warn.
// Сообщение формируется с помощью fmt.Sprintf(template, args...).
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.zapLogger.Warn(fmt.Sprintf(template, args...))
}

// Errorf записывает сообщение уровня error.
// Сообщение формируется с помощью fmt.Sprintf(template, args...).
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.zapLogger.Error(fmt.Sprintf(template, args...))
}
