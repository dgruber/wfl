package jstream_test

import (
	. "github.com/dgruber/wfl/jstream"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func createCoroutineCounter(max *int) JobMap {
	*max = 0
	type counter struct {
		sync.Mutex
		value int
	}
	var c counter
	return func(j *wfl.Job) *wfl.Job {
		c.Lock()
		c.value++
		c.Unlock()
		j.After(time.Millisecond * 100)
		c.Lock()
		if c.value > *max {
			*max = c.value
		}
		c.value--
		c.Unlock()
		return j
	}
}

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

		It("job template must be evaluated before job submission", func() {

			breaker := func(seqLength int) Break {
				seq := seqLength
				return func(t drmaa2interface.JobTemplate) bool {
					if t.JobEnvironment["TASK_ID"] == "2" {
						return false
					}
					seq--
					return seq >= 0
				}
			}
			stream := NewStream(cfg, breaker(10))
			Ω(stream).ShouldNot(BeNil())
			Ω(stream.Error()).Should(BeNil())
			Ω(stream.HasError()).Should(BeFalse())
			jobs := stream.Collect()
			Ω(len(jobs)).Should(BeNumerically("==", 1))
			Ω(jobs[0].Template().JobEnvironment["TASK_ID"]).Should(Equal("1"))

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

		It("should be possible to run all functions in ApplyAsync() concurrently", func() {
			cfg.BufferSize = 1
			stream := NewStream(cfg, NewSequenceBreaker(25))
			Ω(stream).ShouldNot(BeNil())
			Ω(stream.Error()).Should(BeNil())
			Ω(stream.HasError()).Should(BeFalse())

			var max int
			stream.ApplyAsyncN(createCoroutineCounter(&max), 25).Consume()
			Ω(max).Should(BeNumerically("==", 25))

			NewStream(cfg, NewSequenceBreaker(25)).ApplyAsyncN(createCoroutineCounter(&max), 10).Consume()
			Ω(max).Should(BeNumerically("==", 10))

			NewStream(cfg, NewSequenceBreaker(10)).ApplyAsyncN(createCoroutineCounter(&max), 1).Consume()
			Ω(max).Should(BeNumerically("==", 1))

			NewStream(cfg, NewSequenceBreaker(10)).ApplyAsyncN(createCoroutineCounter(&max), 0).Consume()
			Ω(max).Should(BeNumerically("==", 10))
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
				if taskid, _ := strconv.Atoi(j.Template().JobEnvironment["TASK_ID"]); taskid > 50 {
					return true
				}
				return false
			}
			jobs := NewStream(cfg, NewSequenceBreaker(100)).Filter(environmentFilter).Collect()
			Ω(len(jobs)).Should(BeNumerically("==", 50))
			Ω(jobs[0].Template().JobEnvironment["TASK_ID"]).Should(Equal("51"))
			Ω(jobs[49].Template().JobEnvironment["TASK_ID"]).Should(Equal("100"))
		})

		It("should be possible to Tee() a stream", func() {
			cfg.BufferSize = 10
			jobs1, jobs2 := NewStream(cfg, NewSequenceBreaker(1000)).Tee()
			var amount1 int64
			var amount2 int64
			j1 := jobs1.Apply(func(j *wfl.Job) *wfl.Job {
				atomic.AddInt64(&amount1, 1)
				return j
			})
			j2 := jobs2.Apply(func(j *wfl.Job) *wfl.Job {
				atomic.AddInt64(&amount2, 1)
				return j
			})
			syncedStreams := j1.MultiSync(j2)
			Ω(len(syncedStreams)).Should(BeNumerically("==", 2))
			syncedStreams[0].Join(syncedStreams[1])
			Ω(amount1).Should(BeNumerically("==", amount2))
		})

		It("should be possible to MultiSync() a single stream", func() {
			cfg.BufferSize = 10
			jobs1 := NewStream(cfg, NewSequenceBreaker(1000))
			syncedStreams := jobs1.MultiSync()
			Ω(len(syncedStreams)).Should(BeNumerically("==", 1))
			syncedStreams[0].Collect()
		})

		It("should be possible to Merge() multiple streams", func() {
			s1 := NewStream(cfg, NewSequenceBreaker(50))
			s2 := NewStream(cfg, NewSequenceBreaker(120))
			var amount int64
			s1.Merge(s2).Apply(func(j *wfl.Job) *wfl.Job {
				atomic.AddInt64(&amount, 1)
				return j
			}).Synchronize().Consume()
			Ω(amount).Should(BeNumerically("==", 170))
		})

		It("should be possible to CollectN() jobs from a stream", func() {
			s1 := NewStream(cfg, NewSequenceBreaker(500))
			allJobs := make([]*wfl.Job, 0, 500)
			for i := 0; i < 45; i++ {
				jobs := s1.CollectN(11)
				Ω(len(jobs)).Should(BeNumerically("==", 11))
				allJobs = append(allJobs, jobs...)
			}
			jobs := s1.CollectN(11)
			Ω(len(jobs)).Should(BeNumerically("==", 5))

			allJobs = append(allJobs, jobs...)
			for i := range allJobs {
				allJobs[i].Synchronize()
			}
		})

	})

	Context("Standard error cases", func() {

		It("should return errors with broken config", func() {
			config := Config{
				Workflow: nil,
				Template: nil,
			}
			stream := NewStream(config, NewSequenceBreaker(100))
			Ω(stream).ShouldNot(BeNil())
			Ω(stream.Error()).ShouldNot(BeNil())
			Ω(stream.HasError()).Should(BeTrue())

			config = Config{
				Workflow: wfl.NewWorkflow(wfl.NewProcessContext()),
				Template: nil,
			}
			stream = NewStream(config, NewSequenceBreaker(100))
			Ω(stream).ShouldNot(BeNil())
			Ω(stream.Error()).ShouldNot(BeNil())
			Ω(stream.HasError()).Should(BeTrue())
		})

	})

	Context("Regression tests", func() {

		var cfg Config

		BeforeEach(func() {
			cfg.Template = wfl.NewTemplate(drmaa2interface.JobTemplate{
				RemoteCommand: "/bin/sh",
				Args:          []string{"-c", `echo $TASK_ID`},
			}).AddIterator("tasks", wfl.NewEnvSequenceIterator("TASK_ID", 1, 1))
			cfg.Workflow = wfl.NewWorkflow(wfl.NewProcessContext())
			cfg.BufferSize = 1
		})

		It("should call Next() when no breaker is registered", func() {
			stream := NewStream(cfg, nil)
			jobs := stream.CollectN(10)
			Ω(len(jobs)).Should(BeNumerically("==", 10))
			t := jobs[0].Template()
			Ω(t).ShouldNot(BeNil())
			Ω(t.RemoteCommand).Should(Equal("/bin/sh"))
			Ω(t.JobEnvironment["TASK_ID"]).Should(Equal("1"))
			t = jobs[9].Template()
			Ω(t.JobEnvironment["TASK_ID"]).Should(Equal("10"))
		})

	})

})
