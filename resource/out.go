// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Out struct {
	Clientset   versioned.Interface
	ImageWaiter ImageWaiter
}

type ImageWaiter interface {
	Wait(ctx context.Context, writer io.Writer, image *v1alpha1.Image) (string, error)
}

func (o *Out) Out(ctx context.Context, inDir string, src Source, params OutParams, env oc.Environment, log Logger) (oc.Version, oc.Metadata, error) {
	image, err := o.Clientset.KpackV1alpha1().Images(src.Namespace).Get(ctx, src.Image, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, nil, err
	} else if k8serrors.IsNotFound(err) {
		return nil, nil, errors.Errorf("image '%s' in namespace '%s' does not exist. Please create it first.", src.Image, src.Namespace)
	}

	image, err = updateImage(image, inDir, params, log)
	if err != nil {
		return nil, nil, err
	}

	image, err = o.Clientset.KpackV1alpha1().Images(src.Namespace).Update(ctx, image, metav1.UpdateOptions{})
	if err != nil {
		return nil, nil, err
	}

	log.Infof(purple("Waiting on kpack to process update...\n\n"))
	resultingImage, err := o.ImageWaiter.Wait(context.Background(), os.Stderr, image)
	if err != nil {
		return nil, nil, err
	}

	return oc.Version{"image": resultingImage}, nil, nil
}

func updateImage(image *v1alpha1.Image, inDir string, params OutParams, log Logger) (*v1alpha1.Image, error) {
	if params.BlobUrlFile == "" && params.Commitish == "" {
		return nil, errors.Errorf("either commitish or blob_url_file is required")
	}

	switch {
	case params.Commitish != "":
		fileContents, err := ioutil.ReadFile(filepath.Join(inDir, params.Commitish))
		if err != nil {
			return nil, errors.Wrapf(err, "reading commitish: %s", params.Commitish)
		}
		commit := strings.TrimSpace(string(fileContents))

		if image.Spec.Source.Git == nil {
			return nil, errors.Errorf("image '%s' is not configured to use a git source", image.Name)
		}

		log.Infof("Updating image '%s' in namespace '%s'.\nPrevious revision: %s\nNew revision: %s\n\n",
			image.Name, image.Namespace, red(image.Spec.Source.Git.Revision), green(commit))

		image.Spec.Source.Git.Revision = commit
	case params.BlobUrlFile != "":
		fileContents, err := ioutil.ReadFile(filepath.Join(inDir, params.BlobUrlFile))
		if err != nil {
			return nil, errors.Wrapf(err, "reading blobUrl: %s", params.BlobUrlFile)
		}
		blobUrl := strings.TrimSpace(string(fileContents))

		if image.Spec.Source.Blob == nil {
			return nil, errors.Errorf("image '%s' is not configured to use a blob source", image.Name)
		}

		log.Infof("Updating image '%s' in namespace '%s'.\nPrevious blobUrl: %s\nNew blobUrl: %s\n\n",
			image.Name, image.Namespace, red(image.Spec.Source.Blob.URL), green(blobUrl))

		image.Spec.Source.Blob = &corev1alpha1.Blob{
			URL: blobUrl,
		}
	}
	return image, nil
}

var (
	red    = color("\033[1;31m%s\033[0m")
	green  = color("\033[1;32m%s\033[0m")
	purple = color("\033[1;34m%s\033[0m")
)

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
