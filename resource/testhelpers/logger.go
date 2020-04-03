package testhelpers

import (
	"fmt"
	"strings"
)

type Logger struct {
	Out strings.Builder
}

func (l *Logger) Errorf(message string, args ...interface{}) {
	l.Out.WriteString(fmt.Sprintf(message, args) + "\n")
}

func (l *Logger) Warnf(message string, args ...interface{}) {
	l.Out.WriteString(fmt.Sprintf(message, args) + "\n")
}

func (l *Logger) Infof(message string, args ...interface{}) {
	l.Out.WriteString(fmt.Sprintf(message, args) + "\n")
}

func (l *Logger) Debugf(message string, args ...interface{}) {
	l.Out.WriteString(fmt.Sprintf(message, args) + "\n")
}
