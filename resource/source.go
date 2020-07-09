// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"encoding/json"

	oc "github.com/cloudboss/ofcourse/ofcourse"
)

func NewSource(ocSource oc.Source) (Source, error) {
	marshal, err := json.Marshal(ocSource)
	if err != nil {
		return Source{}, err
	}

	source := Source{}
	err = json.Unmarshal(marshal, &source)
	return source, err
}

type Source struct {
	Image     string `json:"image"`
	Namespace string `json:"namespace"`
}
