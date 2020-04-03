package resource

type Logger interface {
	Errorf(message string, args ...interface{})
	Warnf(message string, args ...interface{})
	Infof(message string, args ...interface{})
	Debugf(message string, args ...interface{})
}
