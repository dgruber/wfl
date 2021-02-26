module github.com/dgruber/wfl

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.4
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

require (
	github.com/containerd/containerd v1.4.3 // indirect
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.9
	github.com/dgruber/drmaa2os/pkg/jobtracker/libdrmaa v0.0.0-20210226134924-0c3d0361cded
	github.com/googleapis/gnostic v0.5.4 // indirect
	github.com/mitchellh/copystructure v1.1.1
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/sirupsen/logrus v1.8.0
)

go 1.15
