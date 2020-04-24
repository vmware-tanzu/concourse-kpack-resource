package resource

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	oc "github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Out struct {
	Clientset   versioned.Interface
	ImageWaiter ImageWaiter
}

type ImageWaiter interface {
	Wait(ctx context.Context, image *v1alpha1.Image) (*v1alpha1.Image, error)
}

func (o *Out) Out(inDir string, src Source, params OutParams, env oc.Environment, log Logger) (oc.Version, oc.Metadata, error) {
	fileContents, err := ioutil.ReadFile(filepath.Join(inDir, params.Commitish))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "reading commitsh: %s", params.Commitish)
	}
	commit := strings.TrimSpace(string(fileContents))

	image, err := o.Clientset.BuildV1alpha1().Images(src.Namespace).Get(src.Image, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	if image.Spec.Source.Git == nil {
		return nil, nil, errors.Errorf("image '%s' is not configured to use a git source", image.Name)
	}

	log.Infof("Updating image '%s' in namespace '%s'.\nPrevious revision: %s\nNew revision: %s\n\n",
		image.Name, image.Namespace, red(image.Spec.Source.Git.Revision), green(commit))

	image.Spec.Source.Git.Revision = commit
	image, err = o.Clientset.BuildV1alpha1().Images(src.Namespace).Update(image)
	if err != nil {
		return nil, nil, err
	}

	log.Infof(purple("Waiting on kpack to process update...\n\n"))
	image, err = o.ImageWaiter.Wait(context.Background(), image)
	if err != nil {
		return nil, nil, err
	}

	return oc.Version{"image": image.Status.LatestImage}, nil, nil
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
