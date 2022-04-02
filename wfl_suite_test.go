package wfl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWfl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Wfl Suite")
}
