package jstream

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/mitchellh/copystructure"
	"sync"
)

// JobMap takes a job and returns a job. Input job and
// output job does not need to be the same.
type JobMap func(*wfl.Job) *wfl.Job

// Break is a function which returns true if the job stream has to
// be continued and false if not. It is up to the implementer if
// the job template (which is going to be executed next) is evaluated
// or not.
type Break func(drmaa2interface.JobTemplate) bool

// Filter returns true if the job is ok and should not be filtered out.
type Filter func(*wfl.Job) bool

// NewSequenceBreaker returns a function which stops execution
// of a job stream after _seqLength_ job submissions.
func NewSequenceBreaker(seqLength int) Break {
	seq := seqLength
	return func(t drmaa2interface.JobTemplate) bool {
		seq--
		return seq >= 0
	}
}

// Stream defines a sequence of jobs emitted through a channel.
type Stream struct {
	config Config
	jch    chan *wfl.Job
	err    error
}

// OnError executes the given function in case an error happened
// in the last function which throws errors.
func (g *Stream) OnError(f func(e error)) *Stream {
	if g.err != nil {
		f(g.err)
	}
	return g
}

func (g *Stream) HasError() bool {
	if g.err != nil {
		return true
	}
	return false
}

// Error returns the error of the last operation.
func (g *Stream) Error() error {
	return g.err
}

type Config struct {
	Workflow   *wfl.Workflow
	Template   *wfl.Template
	BufferSize int
}

func checkConfig(cfg Config) error {
	if cfg.Template == nil {
		return errors.New("Template not set")
	}
	if cfg.Workflow == nil {
		return errors.New("Workflow not set")
	}
	if cfg.BufferSize < 0 {
		return errors.New("BufferSize is not allowed to be negative")
	}
	return nil
}

// NewStream creates a new job stream based on the given
// template. The job submission template for each job
// is created by Template.Next(), i.e. defined by the
// registered iterators. The Break function defines
// the end of the stream. If Break is nil an infinite job
// stream is created.
func NewStream(cfg Config, b Break) *Stream {
	if err := checkConfig(cfg); err != nil {
		return &Stream{err: err}
	}
	jobs := make(chan *wfl.Job, cfg.BufferSize)

	go func() {
		for b == nil || b(cfg.Template.Next()) {
			if b == nil {
				cfg.Template.Next()
			}
			jt, _ := copystructure.Copy(cfg.Template.Jt)
			jobs <- wfl.NewJob(cfg.Workflow).RunT(jt.(drmaa2interface.JobTemplate))
		}
		close(jobs)
	}()

	return &Stream{
		jch:    jobs,
		config: cfg,
	}
}

// Tee creates two streams out of one. Note that both streams
// needs to be consumed in parallel otherwise job emission
// will block when the internal buffer of one stream is full.
//
// See also: MultiSync() and Join().
func (g *Stream) Tee() (*Stream, *Stream) {
	jobs1 := make(chan *wfl.Job, g.config.BufferSize)
	jobs2 := make(chan *wfl.Job, g.config.BufferSize)

	go func() {
		for job := range g.jch {
			jobs1 <- job
			jobs2 <- job
		}
		close(jobs1)
		close(jobs2)
	}()

	return &Stream{
			jch:    jobs1,
			config: g.config,
		}, &Stream{
			jch:    jobs2,
			config: g.config,
		}
}

// Join consumes all jobs of two streams of the same length (like streams created
// by Tee()),
func (g *Stream) Join(s *Stream) {
	for range g.jch {
		<-s.jch
	}
}

// MultiSync starts two coroutines which synchronizes the jobs from the
// two streams (waiting until the jobs are finished). It returns two
// synchronized streams which contains only finished jobs in input order.
func (g *Stream) MultiSync(s *Stream) (*Stream, *Stream) {
	outch1 := make(chan *wfl.Job, g.config.BufferSize)
	outch2 := make(chan *wfl.Job, s.config.BufferSize)

	go func() {
		for job := range g.jch {
			outch1 <- job.Synchronize()
		}
		close(outch1)
	}()

	go func() {
		for job := range s.jch {
			outch2 <- job.Synchronize()
		}
		close(outch2)
	}()

	return &Stream{jch: outch1, config: g.config}, &Stream{jch: outch2, config: s.config}
}

// Merge combines the current stream with the given streams into one stream.
// The order in which jobs in the output stream appear is undefined. The order
// within each stream is preserved but the processing order between different
// streams is undefined. If one input stream blocks the other input streams
// are still processed.
func (g *Stream) Merge(s ...*Stream) *Stream {
	jobs := make(chan *wfl.Job, g.config.BufferSize)
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			for job := range g.jch {
				jobs <- job
			}
			wg.Done()
		}()
		for _, is := range s {
			wg.Add(1)
			go func() {
				for job := range is.jch {
					jobs <- job
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(jobs)
	}()
	return &Stream{
		jch:    jobs,
		config: g.config,
	}
}

func (g *Stream) apply(apply JobMap, maxParallel int) *Stream {
	var coroutineControl chan bool

	throttle := maxParallel
	if throttle > 0 { // if negative then do not block at all
		coroutineControl = make(chan bool, throttle)
	}

	outch := make(chan *wfl.Job, g.config.BufferSize)

	go func() {
		var wg sync.WaitGroup
		for job := range g.jch {
			if throttle > 0 {
				coroutineControl <- true // block when coroutineControl buffer is full
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				out := apply(job)
				if out != nil {
					outch <- out
				}
				if throttle > 0 {
					<-coroutineControl
				}
			}()
		}
		wg.Wait()
		close(outch)
	}()

	return &Stream{jch: outch, config: g.config}
}

// Apply returns a newly created job stream. The job stream contains
// jobs which are the result of applying the JobMap function on
// the jobs of the input stream. If the JobMap returns nil no job
// is forwarded.
//
// The JobMap function is applied to the next job only when the
// previous execution is completed.
func (g *Stream) Apply(apply JobMap) *Stream {
	return g.apply(apply, 1)
}

func (g *Stream) ApplyAsync(apply JobMap) *Stream {
	return g.apply(apply, 0)
}

func (g *Stream) ApplyAsyncN(apply JobMap, n int) *Stream {
	return g.apply(apply, n)
}

func (g *Stream) Consume() {
	for range g.jch {
	}
}

// Collect stores all jobs of the stream in an array and
// returns the array. The job stream must be finit and
// small enough to fit in memory.
func (g *Stream) Collect() []*wfl.Job {
	jobs := make([]*wfl.Job, 0, 128)
	for job := range g.jch {
		jobs = append(jobs, job)
	}
	return jobs
}

// CollectN takes the next _size_ jobs out of the job stream
// and returns them as an array. If the stream ends before
// the array is filled, the array is smaller than then given
// size.
func (g *Stream) CollectN(size int) []*wfl.Job {
	jobs := make([]*wfl.Job, 0, size)
	if size <= 0 {
		return jobs
	}
	i := 0
	for job := range g.jch {
		jobs = append(jobs, job)
		i++
		if i >= size {
			break
		}
	}
	return jobs
}

// Synchronize is a non-blocking call which starts a coroutine
// which loop over all jobs in the stream and waits for each
// job until it is finished and then returns the job.
// The newly created output stream contains only finished jobs.
// The order of the output stream is the same as in the incoming
// stream.
func (g *Stream) Synchronize() *Stream {
	outch := make(chan *wfl.Job, g.config.BufferSize)

	go func() {
		for job := range g.jch {
			outch <- job.Synchronize()
		}
		close(outch)
	}()

	return &Stream{jch: outch, config: g.config}
}

// Filter is a non-blocking call which returns a stream
// containing only jobs which are allowed by the filter.
func (g *Stream) Filter(filter Filter) *Stream {
	outch := make(chan *wfl.Job, g.config.BufferSize)

	go func() {
		for job := range g.jch {
			if filter(job) {
				outch <- job
			}
		}
		close(outch)
	}()

	return &Stream{jch: outch, config: g.config}
}

func (g *Stream) JobChannel() chan *wfl.Job {
	return g.jch
}
