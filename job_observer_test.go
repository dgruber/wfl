package wfl_test

import (
	"github.com/dgruber/wfl"

	"github.com/dgruber/drmaa2interface"

	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobObserver", func() {

	Context("Basic Use Cases", func() {

		It("should create the default observer", func() {
			o := wfl.NewDefaultObserver()
			Ω(o.ErrorHandler).ShouldNot(BeNil())
			Ω(o.FailedHandler).ShouldNot(BeNil())
			Ω(o.SuccessHandler).ShouldNot(BeNil())

			Ω(func() { o.ErrorHandler(errors.New("panic")) }).Should(Panic())
			Ω(wfl.NewWorkflow(wfl.NewProcessContext()).Run("sleep", "0").Observe(o)).ShouldNot(BeNil())
		})

		It("should call the ErrorHandler when task submission failed", func() {
			o := wfl.NewDefaultObserver()
			called := false
			o.ErrorHandler = func(e error) { called = true }
			wfl.NewWorkflow(wfl.NewProcessContext()).Run("thisdoesNOTEXIT", "1").Observe(o)
			Ω(called).Should(BeTrue())
		})

		It("should call the FailedHandler when task exited with exit code != 0", func() {
			o := wfl.NewDefaultObserver()
			called := false
			o.FailedHandler = func(j drmaa2interface.Job) { called = true }
			wfl.NewWorkflow(wfl.NewProcessContext()).Run("./test_scripts/exit.sh", "1").Observe(o)
			Ω(called).Should(BeTrue())
		})

		It("should call the SuccessHandler when task finished successfully", func() {
			o := wfl.NewDefaultObserver()
			called := false
			o.SuccessHandler = func(j drmaa2interface.Job) { called = true }
			wfl.NewWorkflow(wfl.NewProcessContext()).Run("sleep", "0").Observe(o)
			Ω(called).Should(BeTrue())
		})

	})

})
