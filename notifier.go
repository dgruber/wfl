package wfl

import ()

type Notifier struct {
	jch chan *Job
}

func NewNotifier() *Notifier {
	return &Notifier{
		jch: make(chan *Job, 16),
	}
}

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
