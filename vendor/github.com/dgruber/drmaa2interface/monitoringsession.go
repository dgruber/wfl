package drmaa2interface

// MonitoringSession interface defines all methods required for
// implementing a DRMAA2 compatible monitoring session.
type MonitoringSession interface {
	CloseMonitoringSession() error
	GetAllJobs(filter JobInfo) ([]Job, error)
	GetAllQueues(names []string) ([]Queue, error)
	GetAllMachines(names []string) ([]Machine, error)
	GetAllReservations() ([]Reservation, error)
}
