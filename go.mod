module github.com/pivotal/concourse-kpack-resource

go 1.14

require (
	github.com/cloudboss/ofcourse v0.2.1
	github.com/google/go-cmp v0.5.1
	github.com/pivotal/kpack v0.1.2
	github.com/pkg/errors v0.9.1
	github.com/sclevine/spec v1.4.0
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.17.5
