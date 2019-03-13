package drmaa2interface

import (
	"time"
)

// JobTemplate defines all fields for a job request defined by the DRMAA2 standard.
type JobTemplate struct {
	Extensible
	Extension         `xml:"-" json:"-"`
	RemoteCommand     string            `json:"remoteCommand"`
	Args              []string          `json:"args"`
	SubmitAsHold      bool              `json:"submitAsHold"`
	ReRunnable        bool              `json:"reRunnable"`
	JobEnvironment    map[string]string `json:"jobEnvironment"`
	WorkingDirectory  string            `json:"workingDirectory"`
	JobCategory       string            `json:"jobCategory"`
	Email             []string          `json:"email"`
	EmailOnStarted    bool              `json:"emailOnStarted"`
	EmailOnTerminated bool              `json:"emailOnTerminated"`
	JobName           string            `json:"jobName"`
	InputPath         string            `json:"inputPath"`
	OutputPath        string            `json:"outputPath"`
	ErrorPath         string            `json:"errorPath"`
	JoinFiles         bool              `json:"joinFiles"`
	ReservationID     string            `json:"reservationID"`
	QueueName         string            `json:"queueName"`
	MinSlots          int64             `json:"minSlots"`
	MaxSlots          int64             `json:"maxSlots"`
	Priority          int64             `json:"priority"`
	CandidateMachines []string          `json:"candidateMachines"`
	MinPhysMemory     int64             `json:"minPhysMemory"`
	MachineOs         string            `json:"machineOs"`
	MachineArch       string            `json:"machineArch"`
	StartTime         time.Time         `json:"startTime"`
	DeadlineTime      time.Time         `json:"deadlineTime"`
	StageInFiles      map[string]string `json:"stageInFiles"`
	StageOutFiles     map[string]string `json:"stageOutFiles"`
	ResourceLimits    map[string]string `json:"resourceLimits"`
	AccountingID      string            `json:"accountingString"`
}
