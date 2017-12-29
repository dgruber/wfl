package jstream_test

import (
	. "github.com/dgruber/wfl/jstream"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"strconv"
)

var _ = Describe("Jstream", func() {

	Context("Standard operations", func() {

		var cfg Config

		BeforeEach(func() {
			cfg.Template = wfl.NewTemplate(drmaa2interface.JobTemplate{
				RemoteCommand: "/bin/sh",
				Args:          []string{"-c", `echo $TASK_ID`},
			}).AddIterator("tasks", wfl.NewEnvSequenceIterator("TASK_ID", 1, 1))

			cfg.Workflow = wfl.NewWorkflow(wfl.NewProcessContext())
			cfg.BufferSize = 2
		})

		It("should be possible to create a stream", func() {
			stream := NewStream(cfg, NewSequenceBreaker(100))
			Ω(stream).ShouldNot(BeNil())
			Ω(stream.Error()).Should(BeNil())
			Ω(stream.HasError()).Should(BeFalse())
			stream.Consume()
		})

		It("should be possible to Collect() all jobs from a stream", func() {
			for i := 100; i < 201; i = i + 100 {
				stream := NewStream(cfg, NewSequenceBreaker(i))
				Ω(stream).ShouldNot(BeNil())
				Ω(stream.Error()).Should(BeNil())
				Ω(stream.HasError()).Should(BeFalse())
				jobs := stream.Collect()
				Ω(len(jobs)).Should(BeNumerically("==", i))
			}
		})

		It("should be possible to Apply() a function on the jobs of a stream", func() {
			stream := NewStream(cfg, NewSequenceBreaker(100))
			Ω(stream).ShouldNot(BeNil())
			Ω(stream.Error()).Should(BeNil())
			Ω(stream.HasError()).Should(BeFalse())

			amount := 0
			counter := func(j *wfl.Job) *wfl.Job {
				amount++
				return j
			}

			stream.Apply(counter).Consume()

			Ω(amount).Should(BeNumerically("==", 100))

			amount = 0
			NewStream(cfg, NewSequenceBreaker(50)).ApplyAsync(counter).Consume()
			Ω(amount).Should(BeNumerically("==", 50))
		})

		It("should be possible to Synchronize() jobs", func() {
			notDone := 0
			isFinished := func(j *wfl.Job) *wfl.Job {
				if j.State() != drmaa2interface.Done {
					notDone++
				}
				return j
			}
			NewStream(cfg, NewSequenceBreaker(100)).Synchronize().Apply(isFinished).Consume()
			Ω(notDone).Should(BeNumerically("==", 0))

		})

		It("should be possible to Filter() a job", func() {
			environmentFilter := func(j *wfl.Job) bool {
				if taskid, _ := strconv.Atoi(j.Template().JobEnvironment["TASK_ID"]); taskid > 51 {
					return true
				}
				return false
			}
			jobs := NewStream(cfg, NewSequenceBreaker(100)).Filter(environmentFilter).Collect()
			Ω(len(jobs)).Should(BeNumerically("==", 50))
		})

	})

})
