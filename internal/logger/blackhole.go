package logger

// Blackhole логгер "черная дыра". Никуда ничего не пишет. Удобно использовать в тестах.
type Blackhole struct{}

func (l *Blackhole) Debugf(template string, args ...interface{}) {}
func (l *Blackhole) Infof(template string, args ...interface{})  {}
func (l *Blackhole) Warnf(template string, args ...interface{})  {}
func (l *Blackhole) Errorf(template string, args ...interface{}) {}
