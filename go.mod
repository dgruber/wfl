module github.com/dgruber/wfl

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.11
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.4
)

require (
	github.com/deepmap/oapi-codegen v1.9.1
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.20
	github.com/go-chi/chi/v5 v5.0.7
	github.com/mitchellh/copystructure v1.1.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/sirupsen/logrus v1.8.1
)

go 1.16
