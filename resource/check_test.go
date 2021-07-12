// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"context"
	"testing"
	"time"

	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/concourse-kpack-resource/resource"
	"github.com/pivotal/concourse-kpack-resource/resource/testhelpers"
)

func TestCheck(t *testing.T) {
	spec.Run(t, "TestCheck", testCheck)
}

func testCheck(t *testing.T, when spec.G, it spec.S) {
	const (
		imageName = "test-image-name"
		namespace = "test-namespace"
	)

	var (
		firstBuildTime = time.Now()
	)

	it("provides the initial version", func() {
		CheckTest{
			Objects: []runtime.Object{
				&v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "1",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime},
					},
					Status: v1alpha1.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:   corev1alpha1.ConditionSucceeded,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530",
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "not-ready-build",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "2",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime.Add(time.Minute)},
					},
					Status: v1alpha1.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:   corev1alpha1.ConditionSucceeded,
									Status: corev1.ConditionUnknown,
								},
							},
						},
					},
				},
			},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: nil,
			ExpectedVersion: []oc.Version{
				map[string]string{
					"image": "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530",
				},
			},
		}.test(t)
	})

	it("does not return builds already checked", func() {
		CheckTest{
			Objects: []runtime.Object{
				&v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "1",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime},
					},
					Status: v1alpha1.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:   corev1alpha1.ConditionSucceeded,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530",
					},
				},
			},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: map[string]string{
				"image": "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530",
			},
			ExpectedVersion: nil,
		}.test(t)
	})

	it("returns the next version after the previous checked version", func() {
		CheckTest{
			Objects: []runtime.Object{
				&v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "1",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime.Add(time.Minute)},
					},
					Status: v1alpha1.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:   corev1alpha1.ConditionSucceeded,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530",
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name-2",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "2",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime.Add(time.Minute)},
					},
					Status: v1alpha1.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:   corev1alpha1.ConditionSucceeded,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "some/image@sha256:4be3b8b101ee62ba005fcb23d2fa76adad27161a6a60f27f8970e81e9c1def69",
					},
				},
			},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: map[string]string{
				"image": "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530",
			},
			ExpectedVersion: []oc.Version{
				map[string]string{
					"image": "some/image@sha256:4be3b8b101ee62ba005fcb23d2fa76adad27161a6a60f27f8970e81e9c1def69",
				},
			},
		}.test(t)
	})

	it("does not return a pervious checked version if builds out of order", func() {
		CheckTest{
			Objects: []runtime.Object{
				&v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name-2",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "2",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime.Add(time.Minute)},
					},
					Status: v1alpha1.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:   corev1alpha1.ConditionSucceeded,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "some/image@sha256:4be3b8b101ee62ba005fcb23d2fa76adad27161a6a60f27f8970e81e9c1def69",
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "1",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime},
					},
					Status: v1alpha1.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:   corev1alpha1.ConditionSucceeded,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530",
					},
				},
			},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: map[string]string{
				"image": "some/image@sha256:4be3b8b101ee62ba005fcb23d2fa76adad27161a6a60f27f8970e81e9c1def69",
			},
			ExpectedVersion: nil,
		}.test(t)
	})
}

type CheckTest struct {
	Objects []runtime.Object
	Source  resource.Source
	Version oc.Version

	ExpectedOutput  string
	ExpectedVersion []oc.Version
}

func (b CheckTest) test(t *testing.T) {
	t.Helper()
	client := fake.NewSimpleClientset(b.Objects...)

	testLog := &testhelpers.Logger{}
	versions, err := resource.Check(client, b.Source, b.Version, nil, testLog, context.TODO())
	require.NoError(t, err)

	assert.Equal(t, b.ExpectedVersion, versions)

	if b.ExpectedOutput != "" {
		assert.Equal(t, b.ExpectedOutput, testLog.Out.String())
	}
}
