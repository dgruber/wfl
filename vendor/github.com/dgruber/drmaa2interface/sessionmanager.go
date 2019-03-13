package drmaa2interface

// SessionManager defines all methods available from a DRMAA2 compatible
// SessionManager implementation. A SessionManager handles creation,
// opening, closing, destruction of DRMAA2 sessions. Sessions can be
// job sessions, monitoring sessions, or optionally reservation sessions.
// It provides generic methods for querying supported optional functionality
// and versioning.
type SessionManager interface {
	CreateJobSession(name, contact string) (JobSession, error)
	CreateReservationSession(sessionName, contact string) (ReservationSession, error)
	OpenMonitoringSession(sessionName string) (MonitoringSession, error)
	OpenJobSession(sessionName string) (JobSession, error)
	OpenReservationSession(name string) (ReservationSession, error)
	DestroyJobSession(sessionName string) error
	DestroyReservationSession(sessionName string) error
	GetJobSessionNames() ([]string, error)
	GetReservationSessionNames() ([]string, error)
	GetDrmsName() (string, error)
	GetDrmsVersion() (Version, error)
	Supports(Capability) bool
	RegisterEventNotification() (EventChannel, error)
}
