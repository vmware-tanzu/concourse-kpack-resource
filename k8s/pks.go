// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func pksSetup(source *PKSSource) (*rest.Config, error) {
	data := url.Values{
		"client_id":     []string{"pks_cluster_client"},
		"client_secret": []string{""},
		"grant_type":    []string{"password"},
		"username":      []string{source.Username},
		"password":      []string{source.Password},
	}.Encode()

	req, err := http.NewRequest(http.MethodPost, source.Api+"/oauth/token", strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get oidc token")
	}

	type tokenResponse struct {
		IdToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
	}

	var token tokenResponse
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(clientcmdapi.Config{}, &clientcmd.ConfigOverrides{
		AuthInfo: clientcmdapi.AuthInfo{
			AuthProvider: &clientcmdapi.AuthProviderConfig{
				Name: "oidc",
				Config: map[string]string{
					"client-id":             "pks_cluster_client",
					"cluster_client_secret": "",
					"id-token":              token.IdToken,
					"idp-issuer-url":        source.Api + "/oauth/token",
					"refresh-token":         token.RefreshToken,
				},
			},
		},
		ClusterInfo: clientcmdapi.Cluster{
			Server:                "https://" + source.Cluster + ":8443",
			InsecureSkipTLSVerify: source.Insecure,
		},
	}).ClientConfig()
}
