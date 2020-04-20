package resource

import (
	"context"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*Todo: This is workaround for a bug in kpack

kpack will report READY=True while a source resolver is waiting to resolve.
This status is incorrect and will make the ImageWaiter exit too soon.
This delay prevents that from occurring.

*/
type DelayedImageWaiter struct {
	KpackClient versioned.Interface
	ImageWaiter ImageWaiter
}

func (d DelayedImageWaiter) Wait(ctx context.Context, originalImage *v1alpha1.Image) (*v1alpha1.Image, error) {
	time.Sleep(5 * time.Second)

	//fetch current version of image to skip image in that falsely reported ready
	imageAfterDelay, err := d.KpackClient.BuildV1alpha1().Images(originalImage.Namespace).Get(originalImage.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	//fix build counter for logs to work
	imageAfterDelay.Status.BuildCounter = originalImage.Status.BuildCounter
	return d.ImageWaiter.Wait(ctx, imageAfterDelay)
}
