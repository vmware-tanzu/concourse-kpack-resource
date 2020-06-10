package resource_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/concourse-kpack-resource/resource"
	"github.com/pivotal/concourse-kpack-resource/resource/testhelpers"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestIn(t *testing.T) {
	spec.Run(t, "TestIn", testIn)
}

func testIn(t *testing.T, when spec.G, it spec.S) {
	const (
		imageName = "test-image-name"
		namespace = "test-namespace"

		imageVersion = "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530"
	)

	var (
		firstBuildTime = time.Now()
		outDir         string
	)

	it.Before(func() {
		var err error
		outDir, err = ioutil.TempDir("", "in_test")
		require.NoError(t, err)
	})

	it.After(func() {
		os.RemoveAll(outDir)
	})

	it("fetches git metadata and writes the image to file", func() {
		InTest{
			Objects: []runtime.Object{
				&v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name-1",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "1",
						},
						Annotations: map[string]string{
							v1alpha1.BuildReasonAnnotation: "Build1Reason",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime},
					},
					Spec: v1alpha1.BuildSpec{
						Source: v1alpha1.SourceConfig{
							Git: &v1alpha1.Git{
								URL:      "gitUrl",
								Revision: "gitRevision",
							},
						},
					},
					Status: v1alpha1.BuildStatus{
						LatestImage: imageVersion,
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
						LatestImage: "some/image@sha256:buildtoIgnore",
					},
				},
			},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: oc.Version{
				"image": imageVersion,
			},
			OutDir: outDir,
			ExpectedVersion: oc.Version{
				"image": imageVersion,
			},
			ExpectedMetadata: oc.Metadata{
				{Name: "buildNumber", Value: "1"},
				{Name: "buildName", Value: "build-name-1"},
				{Name: "buildReason", Value: "Build1Reason"},
				{Name: "gitCommit", Value: "gitRevision"},
				{Name: "gitUrl", Value: "gitUrl"},
			},
		}.test(t)

		assertFileContents(t, filepath.Join(outDir, "image"), imageVersion)

	})

	it("fetches metadata from the last build", func() {
		InTest{
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
						Annotations: map[string]string{
							v1alpha1.BuildReasonAnnotation: "Build2Reason",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime.Add(time.Minute)},
					},
					Spec: v1alpha1.BuildSpec{
						Source: v1alpha1.SourceConfig{
							Git: &v1alpha1.Git{
								URL:      "gitUrl",
								Revision: "gitRevision",
							},
						},
					},
					Status: v1alpha1.BuildStatus{
						LatestImage: imageVersion,
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name-1",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "1",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime},
					},
					Spec: v1alpha1.BuildSpec{
						Source: v1alpha1.SourceConfig{
							Git: &v1alpha1.Git{
								URL:      "gitUrl to ignore",
								Revision: "gitRevision to ignore",
							},
						},
					},
					Status: v1alpha1.BuildStatus{
						LatestImage: imageVersion,
					},
				},
			},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: oc.Version{
				"image": imageVersion,
			},
			OutDir: outDir,
			ExpectedVersion: oc.Version{
				"image": imageVersion,
			},
			ExpectedMetadata: oc.Metadata{
				{Name: "buildNumber", Value: "2"},
				{Name: "buildName", Value: "build-name-2"},
				{Name: "buildReason", Value: "Build2Reason"},
				{Name: "gitCommit", Value: "gitRevision"},
				{Name: "gitUrl", Value: "gitUrl"},
			},
		}.test(t)
	})

	it("writes an empty metadata if build no longer exists", func() {
		image := "some/image@sha256:07c5121b7bc36783614544bd4a7cd6618dc04b963d926cf6e318268cfead0530"

		InTest{
			Objects: []runtime.Object{},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: oc.Version{
				"image": image,
			},
			OutDir: outDir,
			ExpectedVersion: oc.Version{
				"image": image,
			},
			ExpectedMetadata: nil,
		}.test(t)

		assertFileContents(t, filepath.Join(outDir, "image"), image)
	})

	it("fetches blob metadata", func() {
		InTest{
			Objects: []runtime.Object{
				&v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name-1",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "1",
						},
						Annotations: map[string]string{
							v1alpha1.BuildReasonAnnotation: "Build1Reason",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime},
					},
					Spec: v1alpha1.BuildSpec{
						Source: v1alpha1.SourceConfig{
							Blob: &v1alpha1.Blob{
								URL: "https://some-blob-url.com",
							},
						},
					},
					Status: v1alpha1.BuildStatus{
						LatestImage: imageVersion,
					},
				},
			},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: oc.Version{
				"image": imageVersion,
			},
			OutDir: outDir,
			ExpectedVersion: oc.Version{
				"image": imageVersion,
			},
			ExpectedMetadata: oc.Metadata{
				{Name: "buildNumber", Value: "1"},
				{Name: "buildName", Value: "build-name-1"},
				{Name: "buildReason", Value: "Build1Reason"},
				{Name: "blobUrl", Value: "https://some-blob-url.com"},
			},
		}.test(t)

		assertFileContents(t, filepath.Join(outDir, "image"), imageVersion)

	})

	it("fetches registry image metadata", func() {
		InTest{
			Objects: []runtime.Object{
				&v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
				},
				&v1alpha1.Build{
					ObjectMeta: v1.ObjectMeta{
						Name:      "build-name-1",
						Namespace: namespace,
						Labels: map[string]string{
							v1alpha1.ImageLabel:       imageName,
							v1alpha1.BuildNumberLabel: "1",
						},
						Annotations: map[string]string{
							v1alpha1.BuildReasonAnnotation: "Build1Reason",
						},
						CreationTimestamp: v1.Time{Time: firstBuildTime},
					},
					Spec: v1alpha1.BuildSpec{
						Source: v1alpha1.SourceConfig{
							Registry: &v1alpha1.Registry{
								Image: "some-source-image@sha256:something",
							},
						},
					},
					Status: v1alpha1.BuildStatus{
						LatestImage: imageVersion,
					},
				},
			},
			Source: resource.Source{
				Image:     imageName,
				Namespace: namespace,
			},
			Version: oc.Version{
				"image": imageVersion,
			},
			OutDir: outDir,
			ExpectedVersion: oc.Version{
				"image": imageVersion,
			},
			ExpectedMetadata: oc.Metadata{
				{Name: "buildNumber", Value: "1"},
				{Name: "buildName", Value: "build-name-1"},
				{Name: "buildReason", Value: "Build1Reason"},
				{Name: "sourceImage", Value: "some-source-image@sha256:something"},
			},
		}.test(t)

		assertFileContents(t, filepath.Join(outDir, "image"), imageVersion)

	})
}

type InTest struct {
	Objects    []runtime.Object
	OutDir     string
	Source     resource.Source
	Parameters oc.Params
	Version    oc.Version

	ExpectedOutput   string
	ExpectedVersion  oc.Version
	ExpectedMetadata oc.Metadata
	ExpectError      string
}

func (b InTest) test(t *testing.T) {
	t.Helper()
	client := fake.NewSimpleClientset(b.Objects...)

	testLog := &testhelpers.Logger{}

	in := resource.In{
		Clientset: client,
	}

	version, metadata, err := in.In(b.OutDir, b.Source, b.Parameters, b.Version, nil, testLog)
	if b.ExpectError == "" {
		require.NoError(t, err)
	} else {
		require.Error(t, err, b.ExpectError)
	}

	assert.Equal(t, b.ExpectedVersion, version)
	assert.Equal(t, b.ExpectedMetadata, metadata)

	if b.ExpectedOutput != "" {
		assert.Equal(t, b.ExpectedOutput, testLog.Out.String())
	}
}

func assertFileContents(t *testing.T, path string, expected string) {
	t.Helper()
	fileContents, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expected, string(fileContents))
}
