package k8s

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"io/ioutil"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

const googleApplicationCredsEnv = "GOOGLE_APPLICATION_CREDENTIALS"

func gkeSetup(gke *GKESource) (*restclient.Config, error) {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = file.Write([]byte(gke.JSONKey))
	if err != nil {
		return nil, err
	}

	err = os.Setenv(googleApplicationCredsEnv, file.Name())
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.Load([]byte(gke.Kubeconfig))
	if err != nil {
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
}
