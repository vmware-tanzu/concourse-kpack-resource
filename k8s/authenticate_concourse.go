package k8s

import (
	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pkg/errors"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
)

const (
	apiKey      = "api"
	clusterKey  = "cluster"
	usernameKey = "username"
	passwordKey = "password"
	insecureKey = "insecure"
)

func Authenticate(source oc.Source) (*versioned.Clientset, error) {
	api, err := read(source, apiKey)
	if err != nil {
		return nil, err
	}

	cluster, err := read(source, clusterKey)
	if err != nil {
		return nil, err
	}

	username, err := read(source, usernameKey)
	if err != nil {
		return nil, err
	}

	password, err := read(source, passwordKey)
	if err != nil {
		return nil, err
	}

	insecure, ok := source[insecureKey].(bool)
	if !ok {
		return nil, errors.Errorf("Invalid param %s %s", insecureKey, source[insecureKey])
	}

	return pksLogin(api, cluster, username, password, insecure)
}

func read(source oc.Source, name string) (string, error) {
	server, ok := source[name].(string)
	if !ok {
		return "", errors.Errorf("Invalid param %s %s", name, source[name])
	}
	return server, nil
}
