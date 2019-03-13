package drmaa2interface

// Machine represents a compute instance implementing the
// extension interface.
type Machine struct {
	Extensible
	Extension      `xml:"-" json:"-"`
	Name           string  `json:"name"`
	Available      bool    `json:"available"`
	Sockets        int64   `json:"sockets"`
	CoresPerSocket int64   `json:"coresPerSocket"`
	ThreadsPerCore int64   `json:"threadsPerCore"`
	Load           float64 `json:"load"`
	PhysicalMemory int64   `json:"physicalMemory"`
	VirtualMemory  int64   `json:"virtualMemory"`
	Architecture   CPU     `json:"architecture"`
	OSVersion      Version `json:"osVersion"`
	OS             OS      `json:"os"`
}

// OS is the operating system type.
type OS int

//go:generate stringer -type=OS
const (
	OtherOS OS = iota
	AIX
	BSD
	Linux
	HPUX
	IRIX
	MacOS
	SunOS
	TRU64
	UnixWare
	Win
	WinNT
)
