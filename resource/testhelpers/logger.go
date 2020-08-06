// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"fmt"
	"strings"
)

type Logger struct {
	Out strings.Builder
}

func (l *Logger) Infof(message string, args ...interface{}) {
	l.Out.WriteString(fmt.Sprintf(message, args...) + "\n")
}

func (l *Logger) Debugf(message string, args ...interface{}) {
	l.Out.WriteString(fmt.Sprintf(message, args...) + "\n")
}
