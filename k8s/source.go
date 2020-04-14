package k8s

import (
	"encoding/json"
	oc "github.com/cloudboss/ofcourse/ofcourse"
)

func NewSource(logger *oc.Logger, ocSource oc.Source) (Source, error) {
	marshal, err := json.Marshal(ocSource)
	if err != nil {
		return Source{}, err
	}

	logger.Infof(string(marshal))

	source := Source{}
	err = json.Unmarshal(marshal, &source)
	return source, err
}

type Source struct {
	PKS *PKSSource `json:"pks,omitempty"`
	GKE *GKESource `json:"gke,omitempty"`
}

type PKSSource struct {
	Api      string `json:"api"`
	Cluster  string `json:"cluster"`
	Insecure bool   `json:"insecure"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type GKESource struct {
	Kubeconfig string `json:"kubeconfig"`
	JSONKey    string `json:"json_key"`
}
