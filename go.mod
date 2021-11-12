module github.com/dgruber/wfl

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.7
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.4
	k8s.io/api => k8s.io/api v0.20.2
)

require (
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.15
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/mitchellh/copystructure v1.1.1
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20211109214657-ef0fda0de508 // indirect
	golang.org/x/sys v0.0.0-20211109184856-51b60fd695b3 // indirect
	golang.org/x/text v0.3.7 // indirect
)

go 1.16
