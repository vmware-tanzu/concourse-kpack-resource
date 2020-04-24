package resource

type Logger interface {
	Infof(message string, args ...interface{})
	Debugf(message string, args ...interface{})
}
