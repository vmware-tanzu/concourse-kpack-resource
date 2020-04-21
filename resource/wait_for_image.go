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
	"k8s.io/apimachinery/pkg/watch"
	watchTools "k8s.io/client-go/tools/watch"
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var nextBuild = int(image.Status.BuildCounter) + 1
	go w.logTailer.Tail(ctx, os.Stderr, image.Name, strconv.Itoa(nextBuild), image.Namespace)

	if done, _ := imageInTerminalState(watch.Event{Object: image}); done {
		return image, nil
	}

	event, err := watchTools.Until(ctx,
		image.ResourceVersion,
		watchOnlyOneImage{kpackClient: w.KpackClient, image: image},
		filterErrors(imageInTerminalState))
	if err != nil {
		return nil, err
	}

	image, ok := event.Object.(*v1alpha1.Image)
	if !ok {
		return nil, errors.New("unexpected object received")
	}

	if image.Status.GetCondition(corev1alpha1.ConditionReady).IsFalse() {
		return nil, errors.Errorf("update to image %s failed", image.Name)
	}

	return image, nil
}

func imageInTerminalState(event watch.Event) (bool, error) {
	image, ok := event.Object.(*v1alpha1.Image)
	if !ok {
		return false, errors.New("unexpected object received")
	}

	if image.Status.ObservedGeneration != image.Generation {
		return false, nil
	}

	return !image.Status.GetCondition(corev1alpha1.ConditionReady).IsUnknown(), nil
}

func filterErrors(condition watchTools.ConditionFunc) watchTools.ConditionFunc {
	return func(event watch.Event) (bool, error) {
		if event.Type == watch.Error {
			return false, errors.Errorf("error on watch %+v", event.Object)
		}

		return condition(event)
	}
}

type watchOnlyOneImage struct {
	kpackClient versioned.Interface
	image       *v1alpha1.Image
}

func (w watchOnlyOneImage) Watch(options v1.ListOptions) (watch.Interface, error) {
	options.FieldSelector = fmt.Sprintf("metadata.name=%s", w.image.Name)
	return w.kpackClient.BuildV1alpha1().Images(w.image.Namespace).Watch(options)
}
