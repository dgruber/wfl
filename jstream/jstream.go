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

// Filter returns true if the job filter matches (it should be filtered out).
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

type Stream struct {
	config Config
	jch    chan *wfl.Job
	err    error
}

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
				outch <- apply(job)
				if throttle > 0 {
					<-coroutineControl
				}
			}()
		}
		wg.Wait()
		close(outch)
	}()

	return &Stream{jch: outch}
}

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

func (g *Stream) Collect() []*wfl.Job {
	jobs := make([]*wfl.Job, 0, 128)
	for job := range g.jch {
		jobs = append(jobs, job)
	}
	return jobs
}

func (g *Stream) Synchronize() *Stream {
	outch := make(chan *wfl.Job, g.config.BufferSize)

	go func() {
		for job := range g.jch {
			outch <- job.Synchronize()
		}
		close(outch)
	}()

	return &Stream{jch: outch}
}

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

	return &Stream{jch: outch}
}

func (g *Stream) JobChannel() chan *wfl.Job {
	return g.jch
}
