package drmaa2os

import (
	"github.com/dgruber/drmaa2interface"
)

type ReservationSession struct {
}

func (rs *ReservationSession) Close() error {
	return nil
}

func (rs *ReservationSession) GetContact() (string, error) {
	return "", nil
}

func (rs *ReservationSession) GetSessionName() (string, error) {
	return "", nil
}

func (rs *ReservationSession) GetReservation(string) (drmaa2interface.Reservation, error) {
	return nil, nil
}

func (rs *ReservationSession) RequestReservation(template drmaa2interface.ReservationTemplate) (drmaa2interface.Reservation, error) {
	return nil, nil
}

func (rs *ReservationSession) GetReservations() ([]Reservation, error) {
	return nil, nil
}
