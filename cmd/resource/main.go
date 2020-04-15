package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/cloudboss/ofcourse/ofcourse"
	"github.com/pivotal/kpack/pkg/logs"

	"github.com/pivotal/concourse-kpack-resource/k8s"
	"github.com/pivotal/concourse-kpack-resource/resource"
)

func main() {
	switch filepath.Base(os.Args[0]) {
	case "check":
		ofcourse.Check(&concourseResource{})
	case "in":
		ofcourse.In(&concourseResource{})
	case "out":
		ofcourse.Out(&concourseResource{})
	default:
		log.Fatalf("invalid args %s", os.Args)
	}
}

type concourseResource struct{}

func (concourseResource) Check(ofcourseSource ofcourse.Source, version ofcourse.Version, env ofcourse.Environment, logger *ofcourse.Logger) ([]ofcourse.Version, error) {
	k8sSource, err := k8s.NewSource(ofcourseSource)
	if err != nil {
		return nil, err
	}

	clientSet, _, err := k8s.Authenticate(k8sSource)
	if err != nil {
		return nil, err
	}

	source, err := resource.NewSource(ofcourseSource)
	if err != nil {
		return nil, err
	}

	return resource.Check(clientSet, source, version, env, logger)
}

func (concourseResource) In(outputDirectory string, source ofcourse.Source, params ofcourse.Params, version ofcourse.Version, env ofcourse.Environment, logger *ofcourse.Logger) (ofcourse.Version, ofcourse.Metadata, error) {

	return version, nil, nil
}

func (concourseResource) Out(inDir string, ofcourseSource ofcourse.Source, params ofcourse.Params, env ofcourse.Environment, logger *ofcourse.Logger) (ofcourse.Version, ofcourse.Metadata, error) {
	k8sSource, err := k8s.NewSource(ofcourseSource)
	if err != nil {
		return nil, nil, err
	}

	clientSet, k8sClient, err := k8s.Authenticate(k8sSource)
	if err != nil {
		return nil, nil, err
	}

	source, err := resource.NewSource(ofcourseSource)
	if err != nil {
		return nil, nil, err
	}

	outParams, err := resource.NewOutParams(params)
	if err != nil {
		return nil, nil, err
	}

	return (&resource.Out{
		Clientset:   clientSet,
		ImageWaiter: resource.DelayedImageWaiter{resource.NewImageWaiter(clientSet, logs.NewBuildLogsClient(k8sClient))},
	}).Out(inDir, source, outParams, env, logger)
}
