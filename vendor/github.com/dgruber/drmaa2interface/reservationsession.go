package drmaa2interface

import (
	"time"
)

// ReservationInfo contains all details of the current state of
// a resource reservation.
type ReservationInfo struct {
	Extensible
	Extension            `xml:"-" json:"-"`
	ReservationID        string    `json:"reservationID"`
	ReservationName      string    `json:"reservationName"`
	ReservationStartTime time.Time `json:"reservationStartTime"`
	ReservationEndTime   time.Time `json:"reservationEndTime"`
	ACL                  []string  `json:"acl"`
	ReservedSlots        int64     `json:"reservedSlots"`
	ReservedMachines     []string  `json:"reservedMachines"`
}

// ReservationTemplate contains ressource requests for a
// resource reservation.
type ReservationTemplate struct {
	Extensible
	Extension         `xml:"-" json:"-"`
	Name              string        `json:"name"`
	StartTime         time.Time     `json:"startTime"`
	EndTime           time.Time     `json:"endTime"`
	Duration          time.Duration `json:"duration"`
	MinSlots          int64         `json:"minSlots"`
	MaxSlots          int64         `json:"maxSlots"`
	JobCategory       string        `json:"jobCategory"`
	UsersACL          []string      `json:"userACL"`
	CandidateMachines []string      `json:"candidateMachines"`
	MinPhysMemory     int64         `json:"minPhysMemory"`
	MachineOs         string        `json:"machineOs"`
	MachineArch       string        `json:"machineArch"`
}

// Reservation implements all methods required to be
// a valid DRMAA2 compatible reservation, created by a
// ReservationSession.
type Reservation interface {
	GetID() (string, error)
	GetSessionName() (string, error)
	GetTemplate() (ReservationTemplate, error)
	GetInfo() (ReservationInfo, error)
	Terminate() error
}

// ReservationSession provides all methods required for a DRMAA2
// compatible reservation session.
type ReservationSession interface {
	Close() error
	GetContact() (string, error)
	GetSessionName() (string, error)
	GetReservation(string) (Reservation, error)
	RequestReservation(ReservationTemplate) (Reservation, error)
	GetReservations() ([]Reservation, error)
}
