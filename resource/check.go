// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"fmt"
	"sort"

	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Check(clientset versioned.Interface, source Source, version oc.Version, env oc.Environment, logger Logger) ([]oc.Version, error) {
	buildList, err := clientset.BuildV1alpha1().Builds(source.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", v1alpha1.ImageLabel, source.Image),
	})
	if err != nil {
		return nil, err
	}

	builds := filterBuilds(buildList.Items)
	index, _ := indexOfBuild(builds, version)
	builds = builds[index+1:]

	var versions []oc.Version
	for _, build := range builds {
		if build.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsTrue() {
			versions = append(versions, map[string]string{
				"image": build.Status.LatestImage,
			})
		}
	}

	return versions, nil
}

func filterBuilds(items []v1alpha1.Build) []v1alpha1.Build {
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreationTimestamp.Before(&items[j].CreationTimestamp)
	})
	return items
}

func indexOfBuild(items []v1alpha1.Build, version oc.Version) (int, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		build := items[i]
		if build.Status.LatestImage != "" && build.Status.LatestImage == version["image"] {
			return i, true
		}
	}
	return -1, false
}
