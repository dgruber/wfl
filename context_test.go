package wfl_test

import (
	"github.com/dgruber/wfl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
)

var _ = Describe("Context", func() {

	Context("DRMAA2 Context", func() {

		BeforeEach(func() {
			os.Remove("tmp.db")
		})

		It("should be possible to create a drmaa2 process context", func() {
			ctx := wfl.NewProcessContext()
			err := ctx.Error()
			Ω(err).Should(BeNil())
			Ω(ctx).ShouldNot(BeNil())
		})

		It("should be possible to create a drmaa2 docker context", func() {
			ctx := wfl.NewDockerContext("golang:latest", "tmp.db")
			err := ctx.Error()
			Ω(err).Should(BeNil())
			Ω(ctx).ShouldNot(BeNil())
		})

		It("should be possible to create a cloud foundry tasks context", func() {
			ctx := wfl.NewCloudFoundryContext("https://api.run.pivotal.io", "test", "test", "tmp.db")
			err := ctx.Error()
			Ω(err).Should(BeNil())
			Ω(ctx).ShouldNot(BeNil())
		})

		It("should be possible to create an empty test context", func() {
			ctx := wfl.DRMAA2SessionManagerContext(nil)
			err := ctx.Error()
			Ω(err).Should(BeNil())
			Ω(ctx).ShouldNot(BeNil())
		})

		It("should be possible to create an raw drmaa2 session manager context", func() {
			ctx := wfl.DRMAA2SessionManagerContext(nil)
			err := ctx.Error()
			Ω(err).Should(BeNil())
			Ω(ctx).ShouldNot(BeNil())
		})

		It("should execute a function when an error in context creation happened", func() {
			ctx := wfl.ErrorTestContext()
			var e error
			ctx.OnError(func(err error) {
				e = err
			})
			Ω(e).ShouldNot(BeNil())
			err := ctx.Error()
			Ω(err).ShouldNot(BeNil())
			Ω(ctx).ShouldNot(BeNil())
		})

	})

})
