package drmaa2interface

// Capability is a type which represents the availability of optional
// functionality of the DRMAA2 implementation. Optional functionality
// is defined by the DRMAA2 standard but not mandatory to implement.
type Capability int

//go:generate stringer -type=Capability
const (
	AdvanceReservation Capability = iota
	ReserveSlots
	Callback
	BulkJobsMaxParallel
	JtEmail
	JtStaging
	JtDeadline
	JtMaxSlots
	JtAccountingID
	RtStartNow
	RtDuration
	RtMachineOS
	RtMachineArch
)
