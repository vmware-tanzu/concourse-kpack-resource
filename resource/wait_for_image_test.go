package resource

import (
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	clientgotesting "k8s.io/client-go/testing"
)

func TestWatch(t *testing.T) {
	spec.Run(t, "Image Waiter", waitTest)
}

func waitTest(t *testing.T, when spec.G, it spec.S) {
	var (
		fakeLogTailer = &fakeLogTailer{}

		testWatcher = &TestWatcher{
			initialResourceVersion: 1,
			stopped:                false,
			events:                 make(chan watch.Event, 100),
		}

		nextBuild    = 11
		imageToWatch = &v1alpha1.Image{
			ObjectMeta: v1.ObjectMeta{
				Name:            "some-name",
				Namespace:       "some-namespace",
				ResourceVersion: "1",
			},
			Status: v1alpha1.ImageStatus{
				BuildCounter: int64(nextBuild - 1),
			},
		}
		clientset   = fake.NewSimpleClientset()
		imageWaiter = NewImageWaiter(clientset, fakeLogTailer)
	)

	it.Before(func() {
		clientset.PrependWatchReactor("images", func(action clientgotesting.Action) (handled bool, ret watch.Interface, err error) {
			namespace := action.GetNamespace()
			if namespace != imageToWatch.Namespace {
				t.Error("Unexpected namespace watch")
				return false, nil, nil
			}

			watchAction := action.(clientgotesting.WatchAction)
			if watchAction.GetWatchRestrictions().ResourceVersion != imageToWatch.ResourceVersion {
				t.Error("Expected watch on resource version")
				return false, nil, nil
			}

			return true, testWatcher, nil
		})
	})

	it.After(func() {
		assert.Eventually(t, testWatcher.isStopped, time.Second, time.Millisecond)

		assert.True(t, testWatcher.stopped)
		assert.Eventually(t, fakeLogTailer.IsDone, time.Second, time.Millisecond)
		assert.True(t, fakeLogTailer.done)
		assert.Equal(t, []interface{}{os.Stderr, imageToWatch.Name, strconv.Itoa(nextBuild), imageToWatch.Namespace}, fakeLogTailer.args)

		close(testWatcher.events)
	})

	it("returns on image ready and tails logs", func() {
		readyImage := &v1alpha1.Image{
			ObjectMeta: imageToWatch.ObjectMeta,
			Status: v1alpha1.ImageStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
		}

		testWatcher.addEvent(watch.Event{
			Type:   watch.Modified,
			Object: readyImage,
		})

		image, err := imageWaiter.Wait(context.Background(), imageToWatch)
		assert.NoError(t, err)
		assert.Equal(t, readyImage, image)

		assert.True(t, testWatcher.stopped)
		assert.Eventually(t, fakeLogTailer.IsDone, time.Second, time.Millisecond)
		assert.True(t, fakeLogTailer.done)
		assert.Equal(t, []interface{}{os.Stderr, imageToWatch.Name, strconv.Itoa(nextBuild), imageToWatch.Namespace}, fakeLogTailer.args)
	})

	it("only returns when image generation matches observed generation", func() {
		readyImage := &v1alpha1.Image{
			ObjectMeta: imageToWatch.ObjectMeta,
			Status: v1alpha1.ImageStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
		}

		notMatchingGeneration := readyImage.DeepCopy()
		notMatchingGeneration.Generation = 10
		notMatchingGeneration.Status.ObservedGeneration = 9

		testWatcher.addEvent(watch.Event{
			Type:   watch.Modified,
			Object: notMatchingGeneration,
		})

		matchingGeneration := readyImage.DeepCopy()
		matchingGeneration.Generation = 10
		matchingGeneration.Status.ObservedGeneration = 10

		testWatcher.addEvent(watch.Event{
			Type:   watch.Modified,
			Object: matchingGeneration,
		})

		image, err := imageWaiter.Wait(context.Background(), imageToWatch)
		assert.NoError(t, err)
		assert.Equal(t, matchingGeneration, image)
	})

	it("returns an error if image is Ready False", func() {
		readyImage := &v1alpha1.Image{
			ObjectMeta: imageToWatch.ObjectMeta,
			Status: v1alpha1.ImageStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
		}

		testWatcher.addEvent(watch.Event{Type: watch.Modified, Object: readyImage})

		_, err := imageWaiter.Wait(context.Background(), imageToWatch)
		assert.EqualError(t, err, "update to image some-name failed")
	})

	it("does not return an error if image is Ready False but has not observed generation", func() {
		notReadyNotObservedImage := &v1alpha1.Image{
			ObjectMeta: v1.ObjectMeta{
				Name:       "some-name",
				Namespace:  "some-namespace",
				Generation: 10,
			},
			Status: v1alpha1.ImageStatus{
				Status: corev1alpha1.Status{
					ObservedGeneration: 9,
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
		}
		testWatcher.addEvent(watch.Event{
			Type:   watch.Modified,
			Object: notReadyNotObservedImage,
		})

		readyObservedImage := &v1alpha1.Image{
			ObjectMeta: v1.ObjectMeta{
				Name:       "some-name",
				Namespace:  "some-namespace",
				Generation: 10,
			},
			Status: v1alpha1.ImageStatus{
				Status: corev1alpha1.Status{
					ObservedGeneration: 10,
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
		}
		testWatcher.addEvent(watch.Event{
			Type:   watch.Modified,
			Object: readyObservedImage,
		})

		image, err := imageWaiter.Wait(context.Background(), imageToWatch)
		assert.NoError(t, err)
		assert.Equal(t, readyObservedImage, image)
	})

	it("returns an error on watch error", func() {
		testWatcher.events <- watch.Event{
			Type:   watch.Modified,
			Object: &apierrors.NewInternalError(errors.New("this will cause an error in retry watcher")).ErrStatus,
		}

		_, err := imageWaiter.Wait(context.Background(), imageToWatch)
		require.Error(t, err, "error on watch")
		assert.Contains(t, err.Error(), "error on watch")
	})
}

type TestWatcher struct {
	stopped                bool
	events                 chan watch.Event
	initialResourceVersion int
}

func (t *TestWatcher) addEvent(event watch.Event) {
	t.initialResourceVersion++

	image := event.Object.(*v1alpha1.Image)
	image.ResourceVersion = strconv.Itoa(t.initialResourceVersion)
	t.events <- event
}

func (t *TestWatcher) Stop() {
	t.stopped = true
}

func (t TestWatcher) ResultChan() <-chan watch.Event {
	return t.events
}

func (t *TestWatcher) isStopped() bool {
	return t.stopped
}

type fakeLogTailer struct {
	done bool
	args []interface{}
}

func (f *fakeLogTailer) IsDone() bool {
	return f.done
}

func (f *fakeLogTailer) Tail(context context.Context, writer io.Writer, image, build, namespace string) error {
	f.args = []interface{}{writer, image, build, namespace}
	<-context.Done()
	f.done = true
	return nil
}
