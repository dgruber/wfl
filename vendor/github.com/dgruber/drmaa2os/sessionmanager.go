package drmaa2os

import (
	"code.cloudfoundry.org/lager"
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/slurmcli"
	"github.com/dgruber/drmaa2os/pkg/storage"
)

// SessionType represents the selected resource manager.
type SessionType int

const (
	// DefaultSession handles jobs as processes
	DefaultSession SessionType = iota
	// DockerSession manages Docker containers
	DockerSession
	// CloudFoundrySession manages Cloud Foundry application tasks
	CloudFoundrySession
	// KubernetesSession creates Kubernetes jobs
	KubernetesSession
	// SingularitySession manages Singularity containers
	SingularitySession
	// SlurmSession manages slurm jobs
	SlurmSession
)

// SessionManager allows to create, list, and destroy job, reserveration,
// and monitoring sessions. It also returns holds basic information about
// the resource manager and its capabilities.
type SessionManager struct {
	store       storage.Storer
	log         lager.Logger
	sessionType SessionType
	cf          cfContact
	slurm       *slurmcli.Slurm
}

// NewDefaultSessionManager creates a SessionManager which starts jobs
// as processes.
func NewDefaultSessionManager(dbpath string) (*SessionManager, error) {
	return makeSessionManager(dbpath, DefaultSession)
}

// NewSingularitySessionManager creates a new session manager creating and
// maintaining jobs as Singularity containers.
func NewSingularitySessionManager(dbpath string) (*SessionManager, error) {
	return makeSessionManager(dbpath, SingularitySession)
}

// NewDockerSessionManager creates a SessionManager which maintains jobs as
// Docker containers.
func NewDockerSessionManager(dbpath string) (*SessionManager, error) {
	return makeSessionManager(dbpath, DockerSession)
}

// NewCloudFoundrySessionManager creates a SessionManager which maintains jobs
// as Cloud Foundry tasks.
// addr needs to point to the cloud controller API and username and password
// needs to be set as well.
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

// NewKubernetesSessionManager creates a new session manager which uses
// Kubernetes tasks as execution backend for jobs.
func NewKubernetesSessionManager(dbpath string) (*SessionManager, error) {
	return makeSessionManager(dbpath, KubernetesSession)
}

// NewSlurmSessionManager creates a new session manager which wraps the
// slurm command line for managing jobs.
func NewSlurmSessionManager(dbpath string) (*SessionManager, error) {
	return makeSessionManager(dbpath, SlurmSession)
}

// CreateJobSession creates a new JobSession for managing jobs.
func (sm *SessionManager) CreateJobSession(name, contact string) (drmaa2interface.JobSession, error) {
	if err := sm.create(storage.JobSessionType, name, contact); err != nil {
		return nil, err
	}
	jt, err := sm.newJobTracker(name)
	if err != nil {
		return nil, err
	}
	js := newJobSession(name, []jobtracker.JobTracker{jt})
	return js, nil
}

// CreateReservationSession creates a new ReservationSession.
func (sm *SessionManager) CreateReservationSession(name, contact string) (drmaa2interface.ReservationSession, error) {
	return nil, ErrorUnsupportedOperation
}

// OpenMonitoringSession opens a session for monitoring jobs.
func (sm *SessionManager) OpenMonitoringSession(sessionName string) (drmaa2interface.MonitoringSession, error) {
	return nil, errors.New("(TODO) not implemented")
}

// OpenJobSession creates a new session for managing jobs. The semantic of a job session
// and the job session name depends on the resource manager.
func (sm *SessionManager) OpenJobSession(name string) (drmaa2interface.JobSession, error) {
	if exists := sm.store.Exists(storage.JobSessionType, name); !exists {
		return nil, errors.New("JobSession does not exist")
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

// OpenReservationSession opens a reservation session.
func (sm *SessionManager) OpenReservationSession(name string) (drmaa2interface.ReservationSession, error) {
	return nil, ErrorUnsupportedOperation
}

// DestroyJobSession destroys a job session by name.
func (sm *SessionManager) DestroyJobSession(name string) error {
	return sm.delete(storage.JobSessionType, name)
}

// DestroyReservationSession removes a reservation session.
func (sm *SessionManager) DestroyReservationSession(name string) error {
	return ErrorUnsupportedOperation
}

// GetJobSessionNames returns a list of all job sessions.
func (sm *SessionManager) GetJobSessionNames() ([]string, error) {
	return sm.store.List(storage.JobSessionType)
}

// GetReservationSessionNames returns a list of all reservation sessions.
func (sm *SessionManager) GetReservationSessionNames() ([]string, error) {
	return nil, ErrorUnsupportedOperation
}

// GetDrmsName returns the name of the distributed resource manager.
func (sm *SessionManager) GetDrmsName() (string, error) {
	return "drmaa2os", nil
}

// GetDrmsVersion returns the version of the distributed resource manager.
func (sm *SessionManager) GetDrmsVersion() (drmaa2interface.Version, error) {
	return drmaa2interface.Version{Minor: "0", Major: "1"}, nil
}

// Supports returns true of false of the given Capability is supported by DRMAA2OS.
func (sm *SessionManager) Supports(capability drmaa2interface.Capability) bool {
	return false
}

// RegisterEventNotification creates an event channel which emits events when
// the conditions described in the given notification specification are met.
func (sm *SessionManager) RegisterEventNotification() (drmaa2interface.EventChannel, error) {
	return nil, ErrorUnsupportedOperation
}
