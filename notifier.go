package wfl

type Notifier struct {
	jch chan *Job
}

// NewNotifier creates a job notifier which allows to synchronize
// multiple job workflows executed concurrently in go functions.
// Note that there is an internal buffer of 1024 jobs which causes
// SendJob() to block if the buffer is full.
func NewNotifier() *Notifier {
	return &Notifier{
		jch: make(chan *Job, 1024),
	}
}

// SendJob sends a job to the notifier.
func (n *Notifier) SendJob(job *Job) {
	n.jch <- job
}

// ReceiveJob returns a job sent to the notifier.
func (n *Notifier) ReceiveJob() *Job {
	return <-n.jch
}

// Destroy closes the job channel inside the notfier.
func (n *Notifier) Destroy() {
	close(n.jch)
}

// Notify send the job to a notifier.
func (j *Job) Notify(n *Notifier) *Job {
	n.SendJob(j)
	return j
}
