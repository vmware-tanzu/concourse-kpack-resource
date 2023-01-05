// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/concourse-kpack-resource/resource"
	"github.com/pivotal/concourse-kpack-resource/resource/testhelpers"
)

func TestOut(t *testing.T) {
	spec.Run(t, "TestOut", testOut)
}

func testOut(t *testing.T, when spec.G, it spec.S) {
	var inDir string

	it.Before(func() {
		var err error
		inDir, err = ioutil.TempDir("", "outtest")
		require.NoError(t, err)
	})

	it.After(func() {
		os.RemoveAll(inDir)
	})

	when("updating commit", func() {
		const commit = "new-commit"
		const commitishPath = "some-commit-file"

		var (
			image = &v1alpha1.Image{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: v1alpha1.ImageSpec{
					Source: corev1alpha1.SourceConfig{
						Git: &corev1alpha1.Git{
							URL:      "https://some.git.com",
							Revision: "oldrevision",
						},
					},
				},
			}
		)

		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(inDir, commitishPath), []byte(commit+"\n"), 0644)
			require.NoError(t, err)
		})

		it("updates existing images with commit", func() {
			updatedImage := image.DeepCopy()
			updatedImage.Spec.Source.Git.Revision = commit

			OutTest{
				InDir: inDir,
				Objects: []runtime.Object{
					image,
				},
				Source: resource.Source{
					Image:     image.Name,
					Namespace: image.Namespace,
				},
				Parameters: resource.OutParams{
					Commitish: commitishPath,
				},
				TerminalImage: "some.reg.io/image@sha256:1234567",
				ExpectedOutput: []string{
					"Updating image 'test' in namespace 'test-namespace'",
					"Previous revision", "oldrevision",
					"New revision:", "new-commit",
					"Waiting on kpack to process update...\n",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: updatedImage,
					},
				},
				ExpectedVersion: oc.Version{
					"image": "some.reg.io/image@sha256:1234567",
				},
				ExpectedImageToWaitOn: updatedImage,
			}.test(t)

		})

		it("returns error is image does not have a git source", func() {
			image.Spec.Source.Git = nil
			OutTest{
				InDir: inDir,
				Objects: []runtime.Object{
					image,
				},
				Source: resource.Source{
					Image:     image.Name,
					Namespace: image.Namespace,
				},
				Parameters: resource.OutParams{
					Commitish: commitishPath,
				},
				ExpectError: "image 'test' is not configured to use a git source",
			}.test(t)

		})

		it("returns error is image does not exist", func() {
			OutTest{
				InDir:   inDir,
				Objects: nil,
				Source: resource.Source{
					Image:     "does-not-exist",
					Namespace: "some-namespace",
				},
				Parameters: resource.OutParams{
					Commitish: commitishPath,
				},
				ExpectError: "image 'does-not-exist' in namespace 'some-namespace' does not exist. Please create it first.",
			}.test(t)

		})
	})

	when("updating blob_url", func() {
		const blobUrl = "https://new-blob-url.com"
		const blobUrlPath = "some-blob-url-file"

		var (
			image = &v1alpha1.Image{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: v1alpha1.ImageSpec{
					Source: corev1alpha1.SourceConfig{
						Blob: &corev1alpha1.Blob{
							URL: "https://old-blob-url.com",
						},
					},
				},
			}
		)

		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(inDir, blobUrlPath), []byte(blobUrl+"\n"), 0644)
			require.NoError(t, err)
		})

		it("updates existing images with blobUrl", func() {
			updatedImage := image.DeepCopy()
			updatedImage.Spec.Source.Blob.URL = blobUrl

			OutTest{
				InDir: inDir,
				Objects: []runtime.Object{
					image,
				},
				Source: resource.Source{
					Image:     image.Name,
					Namespace: image.Namespace,
				},
				Parameters: resource.OutParams{
					BlobUrlFile: blobUrlPath,
				},
				TerminalImage: "some.reg.io/image@sha256:1234567",
				ExpectedOutput: []string{
					"Updating image 'test' in namespace 'test-namespace'",
					"Previous blobUrl", "https://old-blob-url.com",
					"New blobUrl:", "https://new-blob-url.com",
					"Waiting on kpack to process update...\n",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: updatedImage,
					},
				},
				ExpectedVersion: oc.Version{
					"image": "some.reg.io/image@sha256:1234567",
				},
				ExpectedImageToWaitOn: updatedImage,
			}.test(t)

		})

		it("returns error is image does not have a blob source", func() {
			image.Spec.Source.Blob = nil
			OutTest{
				InDir: inDir,
				Objects: []runtime.Object{
					image,
				},
				Source: resource.Source{
					Image:     image.Name,
					Namespace: image.Namespace,
				},
				Parameters: resource.OutParams{
					BlobUrlFile: blobUrlPath,
				},
				ExpectError: "image 'test' is not configured to use a blob source",
			}.test(t)
		})
	})

	it("returns error if no put parameter is set", func() {
		image := &v1alpha1.Image{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.ImageSpec{
				Source: corev1alpha1.SourceConfig{
					Blob: &corev1alpha1.Blob{
						URL: "https://old-blob-url.com",
					},
				},
			},
		}
		OutTest{
			InDir: inDir,
			Objects: []runtime.Object{
				image,
			},
			Source: resource.Source{
				Image:     image.Name,
				Namespace: image.Namespace,
			},
			Parameters: resource.OutParams{
				Commitish:   "",
				BlobUrlFile: "",
			},
			ExpectError: "either commitish or blob_url_file is required",
		}.test(t)
	})
}

type OutTest struct {
	Objects       []runtime.Object
	InDir         string
	Source        resource.Source
	Parameters    resource.OutParams
	TerminalImage string
	TerminalError error

	ExpectedOutput        []string
	ExpectedImageToWaitOn *v1alpha1.Image
	ExpectUpdates         []clientgotesting.UpdateActionImpl
	ExpectCreates         []runtime.Object
	ExpectedVersion       oc.Version
	ExpectedMetadata      oc.Metadata
	ExpectError           string
}

func (b OutTest) test(t *testing.T) {
	t.Helper()
	client := fake.NewSimpleClientset(b.Objects...)

	testLog := &testhelpers.Logger{}

	waiter := &TestImageWaiter{
		terminalImage: b.TerminalImage,
		error:         b.TerminalError,
	}
	out := resource.Out{
		Clientset:   client,
		ImageWaiter: waiter,
	}

	version, metadata, err := out.Out(context.TODO(), b.InDir, b.Source, b.Parameters, nil, testLog)
	if b.ExpectError == "" {
		require.NoError(t, err)
	} else {
		require.EqualError(t, err, b.ExpectError)
	}

	assert.Equal(t, b.ExpectedVersion, version)
	assert.Equal(t, b.ExpectedMetadata, metadata)

	testhelpers.TestUpdatesAndCreates(t, client, b.ExpectUpdates, b.ExpectCreates)

	assert.Equal(t, b.ExpectedImageToWaitOn, waiter.waitedOnImage, "unexpected image was waited on")

	for _, o := range b.ExpectedOutput {
		assert.Contains(t, testLog.Out.String(), o)
	}
}

type TestImageWaiter struct {
	waitedOnImage *v1alpha1.Image
	terminalImage string
	error         error
}

func (w *TestImageWaiter) Wait(ctx context.Context, writer io.Writer, image *v1alpha1.Image) (string, error) {
	w.waitedOnImage = image

	if w.error != nil {
		return "", w.error
	}

	return w.terminalImage, nil
}
