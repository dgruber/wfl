package drmaa2interface

// EventChannel sends Notifications about job events. When subscribed
// they need to be consumed in the rate created by the system.
type EventChannel <-chan Notification

// Event specifies the type of the event
type Event int

const (
	// NewState is set when a job state changed
	NewState Event = iota
	// Migrated is set when a job is re-located (to another host for example)
	Migrated
	// AttributeChange means that some job info member is changed
	AttributeChange
)

// Notification is the argument of the callback function
// automatically called for an event.
type Notification struct {
	Evt         Event    `json:"event"`
	JobID       string   `json:"jobID"`
	SessionName string   `json:"sessionName"`
	State       JobState `json:"jobState"`
}
