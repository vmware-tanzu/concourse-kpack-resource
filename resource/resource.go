// Package resource is an implementation of a Concourse resource.
package resource

import (
	oc "github.com/cloudboss/ofcourse/ofcourse"
)

type Resource struct{}

func (r *Resource) Check(source oc.Source, version oc.Version, env oc.Environment, logger *oc.Logger) ([]oc.Version, error) {

	return []oc.Version{}, nil

}

func (r *Resource) In(outputDirectory string, source oc.Source, params oc.Params, version oc.Version, env oc.Environment, logger *oc.Logger) (oc.Version, oc.Metadata, error) {

	return version, nil, nil
}

func (r *Resource) Out(inputDirectory string, source oc.Source, params oc.Params, env oc.Environment, logger *oc.Logger) (oc.Version, oc.Metadata, error) {

	return nil, nil, nil
}
