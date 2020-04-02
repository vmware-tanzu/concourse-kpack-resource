docker_registry = gcr.io/cf-build-service-public/concourse-kpack-resource

pack:
	pack build $(docker_registry) --builder cloudfoundry/cnb:bionic

docker: pack
	docker build -t $(docker_registry) --build-arg "base_image=${docker_registry}" hack

publish: docker
	docker push $(docker_registry)

test:
	go test -v ./...

fmt:
	find . -name '*.go' | while read -r f; do \
		gofmt -w -s "$$f"; \
	done

.DEFAULT_GOAL := docker

.PHONY: go-mod docker-build docker-push docker test fmt
