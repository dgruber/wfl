module github.com/dgruber/wfl

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.8
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.4
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

require (
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.10
	github.com/dgruber/drmaa2os/pkg/jobtracker/libdrmaa v0.0.0-20210226091710-ceb83e9b4fff
	github.com/googleapis/gnostic v0.5.4 // indirect
	github.com/mitchellh/copystructure v1.1.1
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/sirupsen/logrus v1.8.0
	gotest.tools/v3 v3.0.3 // indirect
)

go 1.15
