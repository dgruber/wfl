package wfl_test

import (
	"github.com/dgruber/wfl"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/drmaa2interface"
	"os"
)

var _ = Describe("Workflow", func() {

	AfterSuite(func() {
		os.Remove("tmp.db")
	})

	Context("Create a workflow successfully", func() {
		BeforeEach(func() {
			os.Remove("tmp.db")
		})

		It("should create a workflow with a DRMAA2 OS SessionManager", func() {
			ctx := wfl.NewProcessContext()
			err := ctx.Error()
			Ω(err).Should(BeNil())
			Ω(ctx).ShouldNot(BeNil())
			wf := wfl.NewWorkflow(ctx)
			Ω(wf).ShouldNot(BeNil())
			wf.OnError(func(e error) {
				Fail("no error, hence the function should not be executed")
			})
			Ω(wf.HasError()).ShouldNot(BeTrue())
		})

	})

	Context("A workflow should not be created", func() {
		BeforeEach(func() {
			os.Remove("tmp.db")
		})

		It("should fail to create a workflow with nil context", func() {
			wf := wfl.NewWorkflow(nil)
			Ω(wf).ShouldNot(BeNil())
			Ω(wf.HasError()).Should(BeTrue())
		})

		It("should fail when there is no session manager in context", func() {
			ctx := wfl.DRMAA2SessionManagerContext(nil)
			Ω(ctx).ShouldNot(BeNil())
			wf := wfl.NewWorkflow(ctx)
			Ω(wf.HasError()).Should(BeTrue())
			Ω(wf.Error()).ShouldNot(BeNil())
			called := false
			Ω(wf.OnError(func(e error) { called = true }))
			Ω(called).Should(BeTrue())
		})

	})

	Context("First job of workflow", func() {

		jtemplate := drmaa2interface.JobTemplate{RemoteCommand: "sleep", Args: []string{"0"}}

		BeforeEach(func() {
			os.Remove("tmp.db")
		})

		It("can be empty", func() {
			wf := wfl.NewWorkflow(nil)
			Ω(wf).ShouldNot(BeNil())
			Ω(wf.HasError()).Should(BeTrue())

			job := wfl.EmptyJob()
			Ω(job).ShouldNot(BeNil())
		})

		It("should fail when session manager is empty", func() {
			ctx := wfl.DRMAA2SessionManagerContext(nil)
			Ω(ctx).ShouldNot(BeNil())
			wf := wfl.NewWorkflow(ctx)
			Ω(wf.HasError()).Should(BeTrue())
		})

		It("should run the first job in workflow", func() {
			job := wfl.NewWorkflow(wfl.NewProcessContext()).RunT(jtemplate)
			Ω(job).ShouldNot(BeNil())
			Ω(job.LastError()).Should(BeNil())
		})

	})

	Context("Multiple workflows", func() {
		BeforeEach(func() {
			os.Remove("tmp.db")
		})

		It("should be possible to use two workflows with the same context in parallel", func() {
			ctx := wfl.NewProcessContext()
			err := ctx.Error()

			Ω(err).Should(BeNil())
			Ω(ctx).ShouldNot(BeNil())

			wf := wfl.NewWorkflow(ctx)
			Ω(wf).ShouldNot(BeNil())

			wf2 := wfl.NewWorkflow(ctx)
			Ω(wf2).ShouldNot(BeNil())
		})

	})

})
