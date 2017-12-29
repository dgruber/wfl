package drmaa2os

import (
	"code.cloudfoundry.org/lager"
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/cftracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/dockertracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/simpletracker"
	"github.com/dgruber/drmaa2os/pkg/storage"
	"github.com/dgruber/drmaa2os/pkg/storage/boltstore"
	"os"
)

type SessionType int

const (
	DefaultSession      SessionType = iota // processes
	DockerSession                          // containers
	CloudFoundrySession                    // application tasks
)

type cfContact struct {
	addr     string
	username string
	password string
}

type SessionManager struct {
	store       storage.Storer
	log         lager.Logger
	sessionType SessionType
	cf          cfContact
}

func (sm *SessionManager) newJobTracker(name string) (jobtracker.JobTracker, error) {
	switch sm.sessionType {
	case DefaultSession:
		return simpletracker.New(name), nil
	case DockerSession:
		return dockertracker.New()
	case CloudFoundrySession:
		return cftracker.New(sm.cf.addr, sm.cf.username, sm.cf.password, name)
	}
	return nil, errors.New("unknown job session type")
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

func NewDefaultSessionManager(dbpath string) (*SessionManager, error) {
	return makeSessionManager(dbpath, DefaultSession)
}

func NewDockerSessionManager(dbpath string) (*SessionManager, error) {
	return makeSessionManager(dbpath, DockerSession)
}

func NewCloudFoundrySessionManager(addr, username, password, dbpath string) (*SessionManager, error) {
	sm, err := makeSessionManager(dbpath, CloudFoundrySession)
	if err != nil {
		return sm, err
	}
	sm.cf = cfContact{
		addr:     addr,
		username: username,
		password: password,
	}
	return sm, nil
}

func (sm *SessionManager) logErr(message string) error {
	//sm.log.Error(message)
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

func (sm *SessionManager) CreateJobSession(name, contact string) (drmaa2interface.JobSession, error) {
	if err := sm.create(storage.JobSessionType, name, contact); err != nil {
		return nil, err
	}
	jt, err := sm.newJobTracker(name)
	if err != nil {
		return nil, err
	}
	js := NewJobSession(name, []jobtracker.JobTracker{jt})
	return js, nil
}

func (sm *SessionManager) CreateReservationSession(name, contact string) (drmaa2interface.ReservationSession, error) {
	if err := sm.create(storage.ReservationSessionType, name, contact); err != nil {
		return nil, err
	}
	return nil, nil
}

func (sm *SessionManager) OpenMonitoringSession(sessionName string) (drmaa2interface.MonitoringSession, error) {
	return nil, nil
}

func (sm *SessionManager) open(t storage.KeyType, name string) error {
	if exists := sm.store.Exists(t, name); !exists {
		return errors.New("Session does not exist")
	}
	return nil
}

func (sm *SessionManager) OpenJobSession(name string) (drmaa2interface.JobSession, error) {
	if err := sm.open(storage.JobSessionType, name); err != nil {
		return nil, err
	}
	jt, err := sm.newJobTracker(name)
	if err != nil {
		return nil, err
	}
	js := JobSession{
		name:    name,
		tracker: []jobtracker.JobTracker{jt},
	}
	return &js, nil
}

func (sm *SessionManager) OpenReservationSession(name string) (drmaa2interface.ReservationSession, error) {
	if err := sm.open(storage.ReservationSessionType, name); err != nil {
		return nil, err
	}
	return nil, nil
}

func (sm *SessionManager) DestroyJobSession(name string) error {
	return sm.delete(storage.JobSessionType, name)
}

func (sm *SessionManager) DestroyReservationSession(name string) error {
	return ErrorUnsupportedOperation
}

func (sm *SessionManager) GetJobSessionNames() ([]string, error) {
	return sm.store.List(storage.JobSessionType)
}

func (sm *SessionManager) GetReservationSessionNames() ([]string, error) {
	return sm.store.List(storage.ReservationSessionType)
}

func (sm *SessionManager) GetDrmsName() (string, error) {
	return "drmaa2os", nil
}

func (sm *SessionManager) GetDrmsVersion() (drmaa2interface.Version, error) {
	return drmaa2interface.Version{Minor: "0", Major: "1"}, nil
}

func (sm *SessionManager) Supports(capability drmaa2interface.Capability) bool {
	return false
}

func (sm *SessionManager) RegisterEventNotification() (drmaa2interface.EventChannel, error) {
	return nil, ErrorUnsupportedOperation
}
