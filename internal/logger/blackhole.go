package logger

type Blackhole struct{}

func (l *Blackhole) Debugf(template string, args ...interface{}) {}
func (l *Blackhole) Infof(template string, args ...interface{})  {}
func (l *Blackhole) Warnf(template string, args ...interface{})  {}
func (l *Blackhole) Errorf(template string, args ...interface{}) {}
