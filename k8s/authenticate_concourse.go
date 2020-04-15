package k8s

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
)

func Authenticate(source Source) (versioned.Interface, kubernetes.Interface, error) {
	client, err := restConfig(source)
	if err != nil {
		return nil, nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(client)
	if err != nil {
		return nil, nil, err
	}

	kpackClient, err := versioned.NewForConfig(client)
	return kpackClient, k8sClient, err
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
