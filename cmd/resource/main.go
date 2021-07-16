// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
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

func (concourseResource) Check(ocSource ofcourse.Source, version ofcourse.Version, env ofcourse.Environment, logger *ofcourse.Logger) ([]ofcourse.Version, error) {
	ctx := context.Background()

	k8sSource, err := k8s.NewSource(ocSource)
	if err != nil {
		return nil, err
	}

	clientSet, _, err := k8s.Authenticate(k8sSource)
	if err != nil {
		return nil, err
	}

	source, err := resource.NewSource(ocSource)
	if err != nil {
		return nil, err
	}

	return resource.Check(ctx, clientSet, source, version, env, logger)
}

func (concourseResource) In(outDir string, ocSource ofcourse.Source, params ofcourse.Params, version ofcourse.Version, env ofcourse.Environment, logger *ofcourse.Logger) (ofcourse.Version, ofcourse.Metadata, error) {
	ctx := context.Background()

	k8sSource, err := k8s.NewSource(ocSource)
	if err != nil {
		return nil, nil, err
	}

	clientSet, _, err := k8s.Authenticate(k8sSource)
	if err != nil {
		return nil, nil, err
	}

	source, err := resource.NewSource(ocSource)
	if err != nil {
		return nil, nil, err
	}

	return (&resource.In{
		Clientset: clientSet,
	}).In(ctx, outDir, source, params, version, env, logger)
}

func (concourseResource) Out(inDir string, ofcourseSource ofcourse.Source, params ofcourse.Params, env ofcourse.Environment, logger *ofcourse.Logger) (ofcourse.Version, ofcourse.Metadata, error) {
	ctx := context.Background()

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
		ImageWaiter: logs.NewImageWaiter(clientSet, logs.NewBuildLogsClient(k8sClient)),
	}).Out(ctx, inDir, source, outParams, env, Logger{})
}

type Logger struct {
}

func (Logger) Infof(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}

func (Logger) Debugf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}
