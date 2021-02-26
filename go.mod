module github.com/dgruber/wfl

replace (
    github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
    k8s.io/api => k8s.io/api v0.20.2
    k8s.io/client-go => k8s.io/client-go v0.20.2
    k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
)

require (
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.8
	github.com/dgruber/drmaa2os/pkg/jobtracker/libdrmaa v0.0.0-20201125152403-8f166a429464
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/sirupsen/logrus v1.7.0
)

go 1.15
