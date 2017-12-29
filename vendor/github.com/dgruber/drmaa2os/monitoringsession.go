package drmaa2os

import (
	"github.com/dgruber/drmaa2interface"
)

type MonitoringSession struct {
}

func CloseMonitoringSession() error {
	return nil
}

func GetAllJobs(filter drmaa2interface.JobInfo) ([]Job, error) {
	return nil, nil
}

func GetAllQueues(names []string) ([]drmaa2interface.Queue, error) {
	return nil, nil
}

func GetAllMachines(names []string) ([]drmaa2interface.Machine, error) {
	return nil, nil
}

func GetAllReservations() ([]Reservation, error) {
	return nil, nil
}
