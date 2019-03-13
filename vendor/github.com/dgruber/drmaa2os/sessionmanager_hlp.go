package drmaa2os

import (
	"code.cloudfoundry.org/lager"
	"errors"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/cftracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/dockertracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/kubernetestracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/simpletracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/singularity"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/slurmcli"
	"github.com/dgruber/drmaa2os/pkg/storage"
	"github.com/dgruber/drmaa2os/pkg/storage/boltstore"
	"os"
)

type cfContact struct {
	addr     string
	username string
	password string
}

func (sm *SessionManager) newJobTracker(name string) (jobtracker.JobTracker, error) {
	switch sm.sessionType {
	case DockerSession:
		return dockertracker.New(name)
	case CloudFoundrySession:
		return cftracker.New(sm.cf.addr, sm.cf.username, sm.cf.password, name)
	case KubernetesSession:
		return kubernetestracker.New(name, nil)
	case SingularitySession:
		return singularity.New(name)
	case SlurmSession:
		return slurmcli.New(name, slurmcli.NewSlurm("sbatch",
			"squeue", "scontrol", "scancel", "sacct", true))
	default: // DefaultSession
		return simpletracker.New(name), nil
	}
}

func makeSessionManager(dbpath string, st SessionType) (*SessionManager, error) {
	s := boltstore.NewBoltStore(dbpath)
	if err := s.Init(); err != nil {
		return nil, err
	}
	l := lager.NewLogger("sessionmanager")
	l.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))
	return &SessionManager{store: s, log: l, sessionType: st}, nil
}

func (sm *SessionManager) logErr(message string) error {
	return errors.New(message)
}

func (sm *SessionManager) create(t storage.KeyType, name string, contact string) error {
	if exists := sm.store.Exists(t, name); exists {
		return sm.logErr("Session already exists")
	}
	if contact == "" {
		contact = name
	}
	if err := sm.store.Put(t, name, contact); err != nil {
		return err
	}
	return nil
}

func (sm *SessionManager) delete(t storage.KeyType, name string) error {
	if err := sm.store.Delete(t, name); err != nil {
		return sm.logErr("Error while deleting")
	}
	return nil
}
