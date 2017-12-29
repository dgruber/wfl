package simpletracker

import (
	"fmt"
	"math"
	"sync"
)

type lastJobID struct {
	sync.Mutex
	id int64
}

func (l *lastJobID) Next() int64 {
	l.Lock()
	defer l.Unlock()
	if l.id == math.MaxInt64 {
		l.id = 0
	}
	l.id = l.id + 1
	return l.id
}

func (l *lastJobID) Set(jobid int64) {
	l.Lock()
	defer l.Unlock()
	l.id = jobid
}

func NewJobID() *lastJobID {
	return &lastJobID{}
}

var jobID *lastJobID

func init() {
	jobID = NewJobID()
}

func GetNextJobID() string {
	return fmt.Sprintf("%d", jobID.Next())
}

func SetJobID(jobid int64) {
	jobID.Set(jobid)
}
