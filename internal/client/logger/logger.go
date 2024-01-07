package logger

type Logger interface {
	Errorf(string, ...interface{})
	Debugf(string, ...interface{})
	Warnf(string, ...interface{})
}
