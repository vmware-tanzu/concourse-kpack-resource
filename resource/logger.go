// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

type Logger interface {
	Infof(message string, args ...interface{})
	Debugf(message string, args ...interface{})
}
