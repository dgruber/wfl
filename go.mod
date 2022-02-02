module github.com/dgruber/wfl

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.9
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.4
//k8s.io/api => k8s.io/api v0.20.2
)

require (
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.18
	github.com/mitchellh/copystructure v1.1.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/sys v0.0.0-20211109184856-51b60fd695b3 // indirect
)

go 1.16
