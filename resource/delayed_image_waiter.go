package resource

import (
	"context"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

/*Todo: This is workaround for a bug in kpack

kpack will report READY=True while a source resolver is waiting to resolve.
This status is incorrect and will make the ImageWaiter exit too soon.
This delay prevents that from occurring.

*/
type DelayedImageWaiter struct {
	ImageWaiter ImageWaiter
}

func (d DelayedImageWaiter) Wait(ctx context.Context, image *v1alpha1.Image) (*v1alpha1.Image, error) {
	time.Sleep(5 * time.Second)
	return d.ImageWaiter.Wait(ctx, image)
}
