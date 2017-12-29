package drmaa2os

import (
	"github.com/dgruber/drmaa2interface"
)

type Reservation struct {
}

func (r *Reservation) GetID() (string, error) {
	return "", nil
}

func (r *Reservation) GetSessionName() (string, error) {
	return "", nil
}

func (r *Reservation) GetTemplate() (drmaa2interface.ReservationTemplate, error) {
	return drmaa2interface.ReservationTemplate{}, nil
}

func (r *Reservation) GetInfo() (drmaa2interface.ReservationInfo, error) {
	return drmaa2interface.ReservationInfo{}, nil
}

func (r *Reservation) Terminate() error {
	return nil
}
