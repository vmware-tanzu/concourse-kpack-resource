package k8s

import (
	"errors"
	"k8s.io/client-go/rest"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
)

func Authenticate(source Source) (*versioned.Clientset, error) {
	client, err := restConfig(source)
	if err != nil {
		return nil, err
	}

	return versioned.NewForConfig(client)
}

func restConfig(source Source) (*rest.Config, error) {
	switch {
	case source.PKS != nil:
		return pksSetup(source.PKS)
	case source.GKE != nil:
		return gkeSetup(source.GKE)
	default:
		return nil, errors.New("no valid cluster config provided")
	}
}
