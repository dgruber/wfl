package wfl

import ()

type Notifier struct {
	jch chan *Job
}

// NewNotifier creates a job notifier which allows to synchronize
// multiple job workflows executed concurrently in go functions.
func NewNotifier() *Notifier {
	return &Notifier{
		jch: make(chan *Job, 1024),
	}
}

// SendJob sends a job to the notifier.
func (n *Notifier) SendJob(job *Job) {
	n.jch <- job
}

func (n *Notifier) ReceiveJob() *Job {
	return <-n.jch
}

func (n *Notifier) Destroy() {
	close(n.jch)
}

func (j *Job) Notify(n *Notifier) *Job {
	n.SendJob(j)
	return j
}
