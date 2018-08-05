package sigar

import (
	"errors"
	"net"
	"time"
)

type Sigar interface {
	CollectCpuStats(collectionInterval time.Duration) (<-chan Cpu, chan<- struct{})
	GetLoadAverage() (LoadAverage, error)
	GetMem() (Mem, error)
	GetSwap() (Swap, error)
	GetFileSystemUsage(string) (FileSystemUsage, error)
	GetSystemInfo() (SystemInfo, error)
	GetSystemDistribution() (SystemDistribution, error)
}

var ErrNotImplemented error = errors.New("Collection not implemented for this operating system")

type Cpu struct {
	User    uint64
	Nice    uint64
	Sys     uint64
	Idle    uint64
	Wait    uint64
	Irq     uint64
	SoftIrq uint64
	Stolen  uint64
	Guest   uint64
}

func (cpu *Cpu) Total() uint64 {
	return cpu.User + cpu.Nice + cpu.Sys + cpu.Idle +
		cpu.Wait + cpu.Irq + cpu.SoftIrq + cpu.Stolen + cpu.Guest
}

func (cpu Cpu) Delta(other Cpu) Cpu {
	return Cpu{
		User:    cpu.User - other.User,
		Nice:    cpu.Nice - other.Nice,
		Sys:     cpu.Sys - other.Sys,
		Idle:    cpu.Idle - other.Idle,
		Wait:    cpu.Wait - other.Wait,
		Irq:     cpu.Irq - other.Irq,
		SoftIrq: cpu.SoftIrq - other.SoftIrq,
		Stolen:  cpu.Stolen - other.Stolen,
		Guest:   cpu.Guest - other.Guest,
	}
}

type LoadAverage struct {
	One, Five, Fifteen float64
}

type Uptime struct {
	Length float64
}

type Mem struct {
	Total      uint64
	Used       uint64
	Free       uint64
	ActualFree uint64
	ActualUsed uint64
}

type Swap struct {
	Total uint64
	Used  uint64
	Free  uint64
}

type CpuList struct {
	List []Cpu
}

type FileSystem struct {
	DirName     string
	DevName     string
	TypeName    string
	SysTypeName string
	Options     string
	Flags       uint32
}

type FileSystemList struct {
	List []FileSystem
}

type FileSystemUsage struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Avail     uint64
	Files     uint64
	FreeFiles uint64
}

type NetProtoV4Stats struct {
	IP   IPStats
	ICMP ICMPStats
	TCP  TCPStats
	UDP  UDPStats
}

type NetProtoV6Stats struct {
	IP   IPStats
	ICMP ICMPStats
	TCP  TCPStats
	UDP  UDPStats
}

type IPStats struct {
	InReceives      uint64
	InHdrErrors     uint64
	InAddrErrors    uint64
	ForwDatagrams   uint64
	InDelivers      uint64
	InDiscards      uint64
	InUnknownProtos uint64
	OutRequests     uint64
	OutDiscards     uint64
	OutNoRoutes     uint64
}

type ICMPStats struct {
	InMsgs          uint64
	InErrors        uint64
	InDestUnreachs  uint64
	OutMsgs         uint64
	OutErrors       uint64 // Not reported by snmp6
	OutDestUnreachs uint64
}

type TCPStats struct {
	ActiveOpens  uint64
	PassiveOpens uint64
	AttemptFails uint64
	EstabResets  uint64
	CurrEstab    uint64 // Instantaneous value of currently established connections
	InSegs       uint64
	OutSegs      uint64
	RetransSegs  uint64
	InErrs       uint64
	OutRsts      uint64
}

type UDPStats struct {
	InDatagrams  uint64
	OutDatagrams uint64
	InErrors     uint64
	NoPorts      uint64
	RcvbufErrors uint64 // Not reported by snmp6
	SndbufErrors uint64 // Not reported by snmp6
}

type NetIface struct {
	Name       string
	MTU        uint64
	Mac        string
	LinkStatus string

	SendBytes      uint64
	RecvBytes      uint64
	SendPackets    uint64
	RecvPackets    uint64
	SendCompressed uint64
	RecvCompressed uint64
	RecvMulticast  uint64

	SendErrors     uint64
	RecvErrors     uint64
	SendDropped    uint64
	RecvDropped    uint64
	SendFifoErrors uint64
	RecvFifoErrors uint64

	RecvFramingErrors uint64
	SendCarrier       uint64
	SendCollisions    uint64
}

type NetIfaceList struct {
	List []NetIface
}

type NetConnState int

const (
	ConnStateEstablished = NetConnState(iota + 1)
	ConnStateSynSent
	ConnStateSynRecv
	ConnStateFinWait1
	ConnStateFinWait2
	ConnStateTimeWait
	ConnStateClose
	ConnStateCloseWait
	ConnStateLastAck
	ConnStateListen
	ConnStateClosing
)

type NetConn struct {
	LocalAddr  net.IP
	RemoteAddr net.IP
	LocalPort  uint64
	RemotePort uint64
	SendQueue  uint64
	RecvQueue  uint64
	Status     NetConnState
}

type NetTcpConnList struct {
	List []NetConn
}

type NetUdpConnList struct {
	List []NetConn
}

type NetRawConnList struct {
	List []NetConn
}

type NetTcpV6ConnList struct {
	List []NetConn
}

type NetUdpV6ConnList struct {
	List []NetConn
}

type NetRawV6ConnList struct {
	List []NetConn
}

type ProcList struct {
	List []int
}

type RunState byte

const (
	RunStateSleep   = 'S'
	RunStateRun     = 'R'
	RunStateStop    = 'T'
	RunStateZombie  = 'Z'
	RunStateIdle    = 'D'
	RunStateUnknown = '?'
)

type ProcState struct {
	Name      string
	State     RunState
	Ppid      int
	Tty       int
	Priority  int
	Nice      int
	Processor int
}

type ProcIo struct {
	ReadBytes  uint64
	WriteBytes uint64
	ReadOps    uint64
	WriteOps   uint64
}

type ProcMem struct {
	Size        uint64
	Resident    uint64
	Share       uint64
	MinorFaults uint64
	MajorFaults uint64
	PageFaults  uint64
}

type ProcTime struct {
	StartTime uint64
	User      uint64
	Sys       uint64
	Total     uint64
}

type ProcArgs struct {
	List []string
}

type ProcExe struct {
	Name string
	Cwd  string
	Root string
}

type DiskList struct {
	List map[string]DiskIo
}

type DiskIo struct {
	ReadOps     uint64
	ReadBytes   uint64
	ReadTimeMs  uint64
	WriteOps    uint64
	WriteBytes  uint64
	WriteTimeMs uint64
	IoTimeMs    uint64
}

func (self DiskIo) Delta(other DiskIo) DiskIo {
	return DiskIo{
		ReadOps:     self.ReadOps - other.ReadOps,
		ReadBytes:   self.ReadBytes - other.ReadBytes,
		ReadTimeMs:  self.ReadTimeMs - other.ReadTimeMs,
		WriteOps:    self.WriteOps - other.WriteOps,
		WriteBytes:  self.WriteBytes - other.WriteBytes,
		WriteTimeMs: self.WriteTimeMs - other.WriteTimeMs,
		IoTimeMs:    self.IoTimeMs - other.IoTimeMs,
	}
}

type SystemInfo struct {
	Sysname    string
	Nodename   string
	Release    string
	Version    string
	Machine    string
	Domainname string
}

type SystemDistribution struct {
	Description string
}
