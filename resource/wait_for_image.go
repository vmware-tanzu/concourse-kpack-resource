package resource

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type imageWaiter struct {
	KpackClient versioned.Interface
	logTailer   ImageLogTailer
}

type ImageLogTailer interface {
	Tail(context context.Context, writer io.Writer, image, build, namespace string) error
}

func NewImageWaiter(kpackClient versioned.Interface, logTailer ImageLogTailer) *imageWaiter {
	return &imageWaiter{KpackClient: kpackClient, logTailer: logTailer}
}

func (w *imageWaiter) Wait(ctx context.Context, image *v1alpha1.Image) (*v1alpha1.Image, error) {
	watch, err := w.KpackClient.BuildV1alpha1().Images(image.Namespace).Watch(v1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", image.Name),
	})
	if err != nil {
		return nil, err
	}
	defer watch.Stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var nextBuild = int(image.Status.BuildCounter) + 1
	go w.logTailer.Tail(ctx, os.Stderr, image.Name, strconv.Itoa(nextBuild), image.Namespace)

	for event := range watch.ResultChan() {
		image, ok := event.Object.(*v1alpha1.Image)
		if !ok {
			return nil, errors.New("unexpected object received")
		}

		if image.Status.ObservedGeneration == image.Generation {
			if image.Status.GetCondition(corev1alpha1.ConditionReady).IsTrue() {
				return image, nil
			}

			if image.Status.GetCondition(corev1alpha1.ConditionReady).IsFalse() {
				return nil, errors.Errorf("update to image %s failed", image.Name)
			}
		}
	}
	return nil, errors.New("error waiting for image update to apply")
}
