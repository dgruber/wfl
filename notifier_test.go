package wfl_test

import (
	"github.com/dgruber/wfl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Notifier", func() {

	Context("happy path", func() {

		It("should be possible to create and destroy a notifier", func() {
			notifier := wfl.NewNotifier()
			Ω(notifier).ShouldNot(BeNil())
			notifier.Destroy()
		})

		It("should receive a job when a job is send", func() {
			notifier := wfl.NewNotifier()
			Ω(notifier).ShouldNot(BeNil())
			defer notifier.Destroy()

			notifier.SendJob(wfl.EmptyJob())
			notifier.SendJob(wfl.EmptyJob())

			notifier.ReceiveJob()
			notifier.ReceiveJob()
		})

		It("should forward the job when Notify() is called", func() {
			notifier := wfl.NewNotifier()
			defer notifier.Destroy()
			wfl.NewJob(wfl.NewWorkflow(wfl.NewProcessContext())).Run("sleep", "0").Notify(notifier)
			job := notifier.ReceiveJob()
			Ω(job).ShouldNot(BeNil())
			ji := job.JobInfo()
			Ω(ji).ShouldNot(BeNil())
		})

	})
})
