package jstream_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestJstream(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jstream Suite")
}
