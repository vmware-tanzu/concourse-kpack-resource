// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"fmt"
	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
)

const imageFile = "image"

type In struct {
	Clientset versioned.Interface
}

func (in *In) In(outDir string, source Source, params oc.Params, version oc.Version, env oc.Environment, logger Logger) (oc.Version, oc.Metadata, error) {
	err := ioutil.WriteFile(filepath.Join(outDir, imageFile), []byte(version["image"]), 0644)
	if err != nil {
		return nil, nil, err
	}

	buildList, err := in.Clientset.BuildV1alpha1().Builds(source.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", v1alpha1.ImageLabel, source.Image),
	})
	if err != nil {
		return nil, nil, err
	}

	builds := filterBuilds(buildList.Items)
	index, ok := indexOfBuild(builds, version)
	if !ok {
		return version, nil, nil
	}

	build := builds[index]

	return version,
		append(oc.Metadata{
			{Name: "buildNumber", Value: build.Labels[v1alpha1.BuildNumberLabel]},
			{Name: "buildName", Value: build.Name},
			{Name: "buildReason", Value: build.Annotations[v1alpha1.BuildReasonAnnotation]},
		}, sourceMetadata(build)...), nil
}

func sourceMetadata(build v1alpha1.Build) []oc.NameVal {
	switch {
	case build.Spec.Source.Git != nil:
		return []oc.NameVal{
			{Name: "gitCommit", Value: build.Spec.Source.Git.Revision},
			{Name: "gitUrl", Value: build.Spec.Source.Git.URL},
		}
	case build.Spec.Source.Blob != nil:
		return []oc.NameVal{
			{Name: "blobUrl", Value: build.Spec.Source.Blob.URL},
		}
	case build.Spec.Source.Registry != nil:
		return []oc.NameVal{
			{Name: "sourceImage", Value: build.Spec.Source.Registry.Image},
		}
	default:
		return nil
	}
}
