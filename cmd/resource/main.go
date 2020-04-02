package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/cloudboss/ofcourse/ofcourse"

	"github.com/pivotal/concourse-kpack-resource/resource"
)

func main() {
	switch filepath.Base(os.Args[0]) {
	case "check":
		ofcourse.Check(&resource.Resource{})
	case "in":
		ofcourse.In(&resource.Resource{})
	case "out":
		ofcourse.Out(&resource.Resource{})
	default:
		log.Fatalf("invalid args %s", os.Args)
	}
}
