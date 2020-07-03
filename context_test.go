package wfl_test

import (
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/cloudfoundry"
	"github.com/dgruber/wfl/pkg/context/docker"
	"github.com/dgruber/wfl/pkg/context/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
)

var _ = Describe("Context", func() {

	Context("DRMAA2 Context", func() {

		BeforeEach(func() {
			os.Remove("tmp.db")
		})

		Context("Process Context", func() {
			It("should be possible to create a process context", func() {
				ctx := wfl.NewProcessContext()
				err := ctx.Error()
				Ω(err).Should(BeNil())
				Ω(ctx).ShouldNot(BeNil())
				Ω(ctx.HasError()).Should(BeFalse())
			})
			It("should be possible to create a process context with configuration", func() {
				ctx := wfl.NewProcessContextByCfg(wfl.ProcessConfig{DBFile: "tmp.db"})
				err := ctx.Error()
				Ω(err).Should(BeNil())
				Ω(ctx).ShouldNot(BeNil())
			})
		})

		Context("Docker Context", func() {
			It("should be possible to create a docker context", func() {
				ctx := docker.NewDockerContext()
				err := ctx.Error()
				Ω(err).Should(BeNil())
				Ω(ctx).ShouldNot(BeNil())
			})
			It("should be possible to create a docker context with configuration", func() {
				ctx := docker.NewDockerContextByCfg(docker.Config{DBFile: "tmp.db", DefaultDockerImage: "golang:latest"})
				err := ctx.Error()
				Ω(err).Should(BeNil())
				Ω(ctx).ShouldNot(BeNil())
			})
		})

		Context("Cloud Foundry Context", func() {
			It("should be possible to create a cloud foundry tasks context", func() {
				ctx := cloudfoundry.NewCloudFoundryContext()
				err := ctx.Error()
				Ω(err).Should(BeNil())
				Ω(ctx).ShouldNot(BeNil())
			})
			It("should be possible to create a cloud foundry tasks context with configuration", func() {
				ctx := cloudfoundry.NewCloudFoundryContextByCfg(cloudfoundry.Config{DBFile: "tmp.db"})
				err := ctx.Error()
				Ω(err).Should(BeNil())
				Ω(ctx).ShouldNot(BeNil())
			})
		})

		Context("Kubernetes Context", func() {
			It("should be possible to create a Kubernetes context", func() {
				ctx := kubernetes.NewKubernetesContext()
				Ω(ctx).ShouldNot(BeNil())
			})
		})

		Context("Singularity Context", func() {
			It("should be possible to create a Singularity context", func() {
				ctx := wfl.NewSingularityContext()
				Ω(ctx).ShouldNot(BeNil())
				err := ctx.Error()
				Ω(err).Should(BeNil())
			})
			It("should be possible to create a Singularity context by config", func() {
				ctx := wfl.NewSingularityContextByCfg(wfl.SingularityConfig{
					DefaultImage: "",
					DBFile:       "",
				})
				Ω(ctx).ShouldNot(BeNil())
				err := ctx.Error()
				Ω(err).Should(BeNil())
			})
		})

		Context("Temporary DB file", func() {
			It("should always be a different filename", func() {
				files := make(map[string]interface{}, 100)
				for i := 0; i < 100; i++ {
					file := wfl.TmpFile()
					Ω(files).ShouldNot(ContainElement(file))
					Ω(file).ShouldNot(ContainSubstring("%d"))
					files[file] = file
				}
			})
		})

		Context("Test Contexts", func() {
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

})
