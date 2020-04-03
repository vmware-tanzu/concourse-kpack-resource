package resource

import (
	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
)

func Out(clientset versioned.Interface, inDir string, src oc.Source, par oc.Params, env oc.Environment, log Logger) (oc.Version, oc.Metadata, error) {

	return nil, nil, nil
}
