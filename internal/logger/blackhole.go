package logger

type Blackhole struct{}

func (l *Blackhole) Debug(template string, args ...interface{}) {}
func (l *Blackhole) Info(template string, args ...interface{})  {}
func (l *Blackhole) Error(template string, args ...interface{}) {}
