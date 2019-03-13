package drmaa2interface

// CPU architecture types
type CPU int

//go:generate stringer -type=CPU
const (
	OtherCPU CPU = iota
	Alpha
	ARM
	ARM64
	Cell
	PARISC
	PARISC64
	x86
	x64
	IA64
	MIPS
	MIPS64
	PowerPC
	PowerPC64
	SPARC
	SPARC64
)
