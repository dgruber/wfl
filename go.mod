module github.com/dgruber/wfl

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.7
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.4
	k8s.io/api => k8s.io/api v0.20.2
//k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
//k8s.io/client-go => k8s.io/client-go v0.20.2
)

require (
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.14
	github.com/mitchellh/copystructure v1.1.1
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/sirupsen/logrus v1.8.1
)

go 1.16
