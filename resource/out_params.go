// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"encoding/json"

	oc "github.com/cloudboss/ofcourse/ofcourse"
)

func NewOutParams(ocParams oc.Params) (OutParams, error) {
	marshal, err := json.Marshal(ocParams)
	if err != nil {
		return OutParams{}, err
	}

	outParams := OutParams{}
	err = json.Unmarshal(marshal, &outParams)
	return outParams, err
}

type OutParams struct {
	Commitish   string `json:"commitish,omitempty"`
	BlobUrlFile string `json:"blob_url_file,omitempty"`
}
