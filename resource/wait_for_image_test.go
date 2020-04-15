package resource

import (
	"context"
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
		clientset     = fake.NewSimpleClientset()
		fakeLogTailer = &fakeLogTailer{}
		imageWaiter   = NewImageWaiter(clientset, fakeLogTailer)

		testWatcher = &TestWatcher{
			stopped: false,
		}

		nextBuild    = 11
		imageToWatch = &v1alpha1.Image{
			ObjectMeta: v1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-namespace",
			},
			Status: v1alpha1.ImageStatus{
				BuildCounter: int64(nextBuild - 1),
			},
		}
	)

	it.Before(func() {
		clientset.PrependWatchReactor("images", func(action clientgotesting.Action) (handled bool, ret watch.Interface, err error) {
			namespace := action.GetNamespace()
			if namespace != imageToWatch.Namespace {
				t.Error("Unexpected namespace watch")
				return false, nil, nil
			}

			watchAction := action.(clientgotesting.WatchAction)
			match, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
			if !found {
				t.Error("Expected watch on name")
				return false, nil, nil
			}
			if match != imageToWatch.Name {
				t.Errorf("Expected watch on name: %s", imageToWatch.Name)
				return false, nil, nil
			}

			return true, testWatcher, nil
		})
	})

	it.After(func() {
		assert.True(t, testWatcher.stopped)
		assert.Eventually(t, fakeLogTailer.IsDone, time.Second, time.Millisecond)
		assert.True(t, fakeLogTailer.done)
		assert.Equal(t, []interface{}{os.Stderr, imageToWatch.Name, strconv.Itoa(nextBuild), imageToWatch.Namespace}, fakeLogTailer.args)
	})

	it("returns on image ready and tails logs", func() {
		events := make(chan watch.Event, 2)
		testWatcher.results = events
		defer close(events)

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

		events <- watch.Event{
			Type:   watch.Modified,
			Object: readyImage,
		}

		image, err := imageWaiter.Wait(context.Background(), imageToWatch)
		assert.NoError(t, err)
		assert.Equal(t, readyImage, image)

		assert.True(t, testWatcher.stopped)
		assert.Eventually(t, fakeLogTailer.IsDone, time.Second, time.Millisecond)
		assert.True(t, fakeLogTailer.done)
		assert.Equal(t, []interface{}{os.Stderr, imageToWatch.Name, strconv.Itoa(nextBuild), imageToWatch.Namespace}, fakeLogTailer.args)
	})

	it("only returns when image generation matches observed generation", func() {
		events := make(chan watch.Event, 2)
		testWatcher.results = events
		defer close(events)

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

		events <- watch.Event{
			Type:   watch.Modified,
			Object: notMatchingGeneration,
		}

		matchingGeneration := readyImage.DeepCopy()
		matchingGeneration.Generation = 10
		matchingGeneration.Status.ObservedGeneration = 10

		events <- watch.Event{
			Type:   watch.Modified,
			Object: matchingGeneration,
		}

		image, err := imageWaiter.Wait(context.Background(), imageToWatch)
		assert.NoError(t, err)
		assert.Equal(t, matchingGeneration, image)

		assert.True(t, testWatcher.stopped)
	})

	it("returns an error if image is Ready False", func() {
		events := make(chan watch.Event, 2)
		testWatcher.results = events
		defer close(events)

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

		events <- watch.Event{
			Type:   watch.Modified,
			Object: readyImage,
		}

		_, err := imageWaiter.Wait(context.Background(), imageToWatch)
		assert.Error(t, err, "update to image some-name failed")

		assert.True(t, testWatcher.stopped)
	})

	it("does not return an error if image is Ready False but has not observed generation", func() {
		events := make(chan watch.Event, 2)
		testWatcher.results = events
		defer close(events)

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
		events <- watch.Event{
			Type:   watch.Modified,
			Object: notReadyNotObservedImage,
		}

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
		events <- watch.Event{
			Type:   watch.Modified,
			Object: readyObservedImage,
		}

		image, err := imageWaiter.Wait(context.Background(), imageToWatch)
		assert.NoError(t, err)
		assert.Equal(t, readyObservedImage, image)

		assert.True(t, testWatcher.stopped)
	})

	it("returns error if image watch closes", func() {
		events := make(chan watch.Event, 2)
		testWatcher.results = events
		close(events)

		_, err := imageWaiter.Wait(context.Background(), imageToWatch)
		require.Errorf(t, err, "error waiting for image update to apply")

		assert.True(t, testWatcher.stopped)
	})
}

func setupTestWatcher(t, clientset *fake.Clientset, imageToWatch *v1alpha1.Image, events chan watch.Event) *TestWatcher {
	testWatcher := &TestWatcher{
		results: events,
		stopped: false,
	}

	clientset.PrependWatchReactor("images", func(action clientgotesting.Action) (handled bool, ret watch.Interface, err error) {
		watchAction := action.(clientgotesting.WatchAction)
		match, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
		if !found {
			return false, nil, nil
		}
		if match != imageToWatch.Name {
			return false, nil, nil
		}

		return true, testWatcher, nil
	})
	return testWatcher
}

type TestWatcher struct {
	stopped bool
	results <-chan watch.Event
}

func (t *TestWatcher) Stop() {
	t.stopped = true
}

func (t TestWatcher) ResultChan() <-chan watch.Event {
	return t.results
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
