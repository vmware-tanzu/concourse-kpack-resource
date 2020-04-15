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
	Commitish string `json:"commitish"`
}
