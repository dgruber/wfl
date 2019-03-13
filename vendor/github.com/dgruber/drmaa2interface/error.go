package drmaa2interface

// ErrorID type represents a DRMAA2 standardized error ID which
// is part of the DRMAA2 Error type.
type ErrorID int

//go:generate stringer -type=ErrorID
const (
	Success ErrorID = iota
	DeniedByDrms
	DrmCommunication
	TryLater
	SessionManagement
	Timeout
	Internal
	InvalidArgument
	InvalidSession
	InvalidState
	OutOfResource
	UnsupportedAttribute
	UnsupportedOperation
	ImplementationSpecific
	LastError
)

// Error is a DRMAA2 error type (implements Go Error interface). All errors
// returned by any DRMAA2 method is (and can be casted to) this Error type.
type Error struct {
	// Messages is a dynamically created human readable error description.
	// There is no fixed message catalog defined.
	Message string
	// ID is used for identifying the error type
	ID ErrorID
}

// Error implements the Error interface.
func (ce Error) Error() string {
	return ce.Message
}

// String implements the Stringer interface for an drmaa2.Error
func (ce Error) String() string {
	return ce.Message
}
