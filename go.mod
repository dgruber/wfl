module github.com/dgruber/wfl

replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0

require (
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.6
	github.com/dgruber/drmaa2os/pkg/jobtracker/libdrmaa v0.0.0-20201125152403-8f166a429464
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/sirupsen/logrus v1.7.0
)

go 1.15
