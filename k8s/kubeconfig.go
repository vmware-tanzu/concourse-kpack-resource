// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func kubeConfigSetup(kubeconfig string) (*restclient.Config, error) {

	config, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
}
