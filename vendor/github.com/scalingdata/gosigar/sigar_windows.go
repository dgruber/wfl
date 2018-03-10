// Copyright (c) 2012 VMware, Inc.

package sigar

import (
	"time"

	"github.com/scalingdata/wmi"
)

/*
// We need both of these libraries to list network connections
#cgo LDFLAGS: -liphlpapi -lws2_32

// IPv6 connection lists are only supported on Vista and above
#define _WIN32_WINNT 0x600

#include <stdlib.h>
#include <winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#include <iphlpapi.h>

// Helper methods to create appropriately-sized MIB tables - it's easier to
// malloc them and cast them to the right type than to do the unsafe.Pointer
// casting in Go. Returns NULL if there's an error, otherwise the pointer
// returned needs to be free'd by the caller.
PMIB_TCPTABLE_OWNER_PID getTcpTable(PDWORD err) {
	PMIB_TCPTABLE_OWNER_PID pTable = (MIB_TCPTABLE_OWNER_PID *) malloc(sizeof(MIB_TCPTABLE_OWNER_PID));
	if (pTable == NULL) {
		*err = 1;
		return NULL;
	}
	DWORD size = sizeof(MIB_TCPTABLE_OWNER_PID);
	if ((*err = GetExtendedTcpTable(pTable, &size, FALSE, AF_INET, TCP_TABLE_OWNER_PID_ALL, 0)) != ERROR_INSUFFICIENT_BUFFER) {
		if (*err == NO_ERROR) {
			return pTable;
		}
		free(pTable);
		return NULL;
	}
	free(pTable);
	pTable = (MIB_TCPTABLE_OWNER_PID *) malloc(size);
	if ((*err = GetExtendedTcpTable(pTable, &size, FALSE, AF_INET, TCP_TABLE_OWNER_PID_ALL, 0)) != NO_ERROR) {
		free(pTable);
		return NULL;
	}
	*err = 0;
	return pTable;
}

PMIB_UDPTABLE_OWNER_PID getUdpTable(PDWORD err) {
	PMIB_UDPTABLE_OWNER_PID pTable = (MIB_UDPTABLE_OWNER_PID *) malloc(sizeof(MIB_UDPTABLE_OWNER_PID));
	if (pTable == NULL) {
		*err = 1;
		return NULL;
	}
	DWORD size = sizeof(MIB_UDPTABLE_OWNER_PID);
	if ((*err = GetExtendedUdpTable(pTable, &size, FALSE, AF_INET, UDP_TABLE_OWNER_PID, 0)) != ERROR_INSUFFICIENT_BUFFER) {
		if (*err == NO_ERROR) {
			return pTable;
		}
		free(pTable);
		return NULL;
	}
	free(pTable);
	pTable = (MIB_UDPTABLE_OWNER_PID *) malloc(size);
	if ((*err = GetExtendedUdpTable(pTable, &size, FALSE, AF_INET, UDP_TABLE_OWNER_PID, 0)) != NO_ERROR) {
		free(pTable);
		return NULL;
	}
	*err = 0;
	return pTable;
}

PMIB_TCP6TABLE_OWNER_PID getTcp6Table(PDWORD err) {
	PMIB_TCP6TABLE_OWNER_PID pTable = (MIB_TCP6TABLE_OWNER_PID *) malloc(sizeof(MIB_TCP6TABLE_OWNER_PID));
	if (pTable == NULL) {
		*err = 1;
		return NULL;
	}
	DWORD size = sizeof(MIB_TCP6TABLE_OWNER_PID);
	if ((*err = GetExtendedTcpTable(pTable, &size, FALSE, AF_INET6, TCP_TABLE_OWNER_PID_ALL, 0)) != ERROR_INSUFFICIENT_BUFFER) {
		if (*err == NO_ERROR) {
			return pTable;
		}
		free(pTable);
		return NULL;
	}
	free(pTable);
	pTable = (MIB_TCP6TABLE_OWNER_PID *) malloc(size);
	if ((*err = GetExtendedTcpTable(pTable, &size, FALSE, AF_INET6, TCP_TABLE_OWNER_PID_ALL, 0)) != NO_ERROR) {
		free(pTable);
		return NULL;
	}
	*err = 0;
	return pTable;
}

PMIB_UDP6TABLE_OWNER_PID getUdp6Table(PDWORD err) {
	PMIB_UDP6TABLE_OWNER_PID pTable = (MIB_UDP6TABLE_OWNER_PID *) malloc(sizeof(MIB_UDP6TABLE_OWNER_PID));
	if (pTable == NULL) {
		*err = 1;
		return NULL;
	}
	DWORD size = sizeof(MIB_UDP6TABLE_OWNER_PID);
	if ((*err = GetExtendedUdpTable(pTable, &size, FALSE, AF_INET6, UDP_TABLE_OWNER_PID, 0)) != ERROR_INSUFFICIENT_BUFFER) {
		if (*err == NO_ERROR) {
			return pTable;
		}
		free(pTable);
		return NULL;
	}
	free(pTable);
	pTable = (MIB_UDP6TABLE_OWNER_PID *) malloc(size);
	if ((*err = GetExtendedUdpTable(pTable, &size, FALSE, AF_INET6, UDP_TABLE_OWNER_PID, 0)) != NO_ERROR) {
		free(pTable);
		return NULL;
	}
	*err = 0;
	return pTable;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"strconv"
	"unsafe"
)

// Use package-global wmi client to avoid modifying wmi.DefaultClient
var wmiClient = &wmi.Client{}

func init() {
	// WMI queries will set the zero value for struct items that can't be read,
	// e.g. due to lack of permission
	wmiClient.NonePtrZero = true
}

func (self *LoadAverage) Get() error {
	return notImplemented()
}

func (self *Uptime) Get() error {
	return notImplemented()
}

func (self *Mem) Get() error {
	var statex C.MEMORYSTATUSEX
	statex.dwLength = C.DWORD(unsafe.Sizeof(statex))

	succeeded := C.GlobalMemoryStatusEx(&statex)
	if succeeded == C.FALSE {
		lastError := C.GetLastError()
		return fmt.Errorf("GlobalMemoryStatusEx failed with error: %d", int(lastError))
	}

	self.Total = uint64(statex.ullTotalPhys)
	self.Free = uint64(statex.ullAvailPhys)
	self.Used = self.Total - self.Free
	return nil
}

func (self *Swap) Get() error {
	swapQueries := []string{
		`\memory\committed bytes`,
		`\memory\commit limit`,
	}

	queryResults, err := runRawPdhQueries(swapQueries)
	if err != nil {
		return err
	}
	self.Used = uint64(queryResults[0])
	self.Total = uint64(queryResults[1])
	self.Free = self.Total - self.Used
	return nil
}

func (self *Cpu) Get() error {
	cpuQueries := []string{
		`\processor(_Total)\% idle time`,
		`\processor(_Total)\% user time`,
		`\processor(_Total)\% privileged time`,
		`\processor(_Total)\% interrupt time`,
	}
	queryResults, err := runRawPdhQueries(cpuQueries)
	if err != nil {
		return err
	}

	self.populateFromPdh(queryResults)
	return nil
}

func (self *CpuList) Get() error {
	cpuQueries := []string{
		`\processor(*)\% idle time`,
		`\processor(*)\% user time`,
		`\processor(*)\% privileged time`,
		`\processor(*)\% interrupt time`,
	}
	// Run a PDH query for all CPU metrics
	queryResults, err := runRawPdhArrayQueries(cpuQueries)
	if err != nil {
		return err
	}
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 4
	}
	self.List = make([]Cpu, 0, capacity)
	for cpu, counters := range queryResults {
		index := 0
		if cpu == "_Total" {
			continue
		}

		index, err := strconv.Atoi(cpu)
		if err != nil {
			continue
		}

		// Expand the array to accomodate this CPU id
		for i := len(self.List); i <= index; i++ {
			self.List = append(self.List, Cpu{})
		}

		// Populate the relevant fields
		self.List[index].populateFromPdh(counters)
	}
	return nil
}

func (self *Cpu) populateFromPdh(values []uint64) {
	self.Idle = values[0]
	self.User = values[1]
	self.Sys = values[2]
	self.Irq = values[3]
}

// Get a list of local filesystems
// Does not apply to SMB volumes
func (self *FileSystemList) Get() error {
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 4
	}
	self.List = make([]FileSystem, 0, capacity)

	iter, err := NewWindowsVolumeIterator()
	if err != nil {
		return err
	}

	for iter.Next() {
		volume := iter.Volume()
		self.List = append(self.List, volume)
	}
	iter.Close()

	return iter.Error()
}

func (self *DiskList) Get() error {
	/* Even though these queries are % disk time and ops / sec,
	   we read the raw PDH counter values, not the "cooked" ones.
	   This gives us the underlying number of ticks that would go into
	   computing the cooked metric. */
	diskQueries := []string{
		`\physicaldisk(*)\disk reads/sec`,
		`\physicaldisk(*)\disk read bytes/sec`,
		`\physicaldisk(*)\% disk read time`,
		`\physicaldisk(*)\disk writes/sec`,
		`\physicaldisk(*)\disk write bytes/sec`,
		`\physicaldisk(*)\% disk write time`,
	}

	// Run a PDH query for metrics across all physical disks
	queryResults, err := runRawPdhArrayQueries(diskQueries)
	if err != nil {
		return err
	}

	self.List = make(map[string]DiskIo)
	for disk, counters := range queryResults {
		if disk == "_Total" {
			continue
		}

		self.List[disk] = DiskIo{
			ReadOps:   uint64(counters[0]),
			ReadBytes: uint64(counters[1]),
			// The raw counter for `% disk read time` is measured
			// in 100ns ticks, divide by 10000 to get milliseconds
			ReadTimeMs: uint64(counters[2] / 10000),
			WriteOps:   uint64(counters[3]),
			WriteBytes: uint64(counters[4]),
			// The raw counter for `% disk write time` is measured
			// in 100ns ticks, divide by 10000 to get milliseconds
			WriteTimeMs: uint64(counters[5] / 10000),
			IoTimeMs:    uint64((counters[5] + counters[2]) / 10000),
		}
	}
	return nil
}

// Used by the wmi package to access the Win32_Process WMI class
type Win32_Process struct {
	Name            string
	ProcessId       uint32
	ExecutablePath  string // Requires SeDebugPrivilege, will not be present without this privilege
	ExecutionState  uint16 // Requires SeDebugPrivilege, will not be present without this privilege
	ParentProcessId uint32
	Priority        uint32
	CommandLine     string
	CreationDate    time.Time

	VirtualSize    uint64
	WorkingSetSize uint64
	PageFaults     uint32

	UserModeTime   uint64
	KernelModeTime uint64

	ReadOperationCount  uint64
	ReadTransferCount   uint64
	WriteOperationCount uint64
	WriteTransferCount  uint64
}

// Used by the WMI package
type Win32_PerfFormattedData_PerfProc_Process struct {
	IDProcess             uint32
	PercentPrivilegedTime uint64
	PercentUserTime       uint64
	PercentProcessorTime  uint64
	PageFileBytes         uint64
}

type WindowsRunState int

const (
	WindowsRunStateUnknown = WindowsRunState(iota)
	WindowsRunStateOther
	WindowsRunStateReady
	WindowsRunStateRunning
	WindowsRunStateBlocked
	WindowsRunStateSuspendedBlocked
	WindowsRunStateSuspendedReady
	WindowsRunStateTerminated
	WindowsRunStateStopped
	WindowsRunStateGrowing
)

func convertWindowsRunState(state WindowsRunState) RunState {
	// This mapping may not be exact, see "ExecutionState" at
	// https://msdn.microsoft.com/en-us/library/aa387976(v=vs.85).aspx
	switch WindowsRunState(state) {
	case WindowsRunStateReady:
	case WindowsRunStateRunning:
		return RunStateRun
	case WindowsRunStateBlocked:
	case WindowsRunStateSuspendedBlocked:
		return RunStateIdle
	case WindowsRunStateTerminated:
		return RunStateZombie
	case WindowsRunStateStopped:
		return RunStateStop
	}
	return RunStateUnknown
}

// Helper to convert 100 nanosecond units to milliseconds
func convert100NsUnitsToMillis(value uint64) uint64 {
	return value / 10000
}

func (self *ProcessList) Get() error {
	// Query process list
	var procs []Win32_Process
	whereClause := ""
	query := wmi.CreateQuery(&procs, whereClause)
	err := wmiClient.Query(query, &procs)
	if err != nil {
		return err
	}

	// Query performance class to get percent user/sys time
	var procPerfs []Win32_PerfFormattedData_PerfProc_Process
	whereClause = ""
	query = wmi.CreateQuery(&procPerfs, whereClause)
	err = wmiClient.Query(query, &procPerfs)
	if err != nil {
		return err
	}

	// Index performance data by process ID for easy lookup
	perfLookup := make(map[uint32]Win32_PerfFormattedData_PerfProc_Process)
	for _, procPerf := range procPerfs {
		perfLookup[procPerf.IDProcess] = procPerf
	}

	// Form list of returned Process structs
	processes := make([]Process, 0, len(procs))
	for _, proc := range procs {
		var process Process
		perf := perfLookup[proc.ProcessId]

		// ProcState
		process.ProcState.Name = proc.Name
		process.ProcState.Pid = int(proc.ProcessId)
		process.ProcState.Ppid = int(proc.ParentProcessId)
		process.ProcState.Priority = int(proc.Priority)
		process.ProcState.State = convertWindowsRunState(WindowsRunState(proc.ExecutionState))

		// ProcIo
		process.ProcIo.ReadBytes = proc.ReadTransferCount
		process.ProcIo.ReadOps = proc.ReadOperationCount
		process.ProcIo.WriteBytes = proc.WriteTransferCount
		process.ProcIo.WriteOps = proc.WriteOperationCount

		// ProcMem
		process.ProcMem.Size = proc.VirtualSize
		process.ProcMem.Resident = proc.WorkingSetSize
		process.ProcMem.PageFaults = uint64(proc.PageFaults)
		process.ProcMem.PageFileBytes = perf.PageFileBytes

		// ProcTime
		process.ProcTime.User = convert100NsUnitsToMillis(proc.UserModeTime)
		process.ProcTime.Sys = convert100NsUnitsToMillis(proc.KernelModeTime)
		process.ProcTime.Total = process.ProcTime.User + process.ProcTime.Sys
		// Convert proc.CreationDate to millis
		process.ProcTime.StartTime = uint64(proc.CreationDate.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond)))
		process.ProcTime.PercentUserTime = perf.PercentUserTime
		process.ProcTime.PercentSysTime = perf.PercentPrivilegedTime
		process.ProcTime.PercentTotalTime = perf.PercentProcessorTime

		// ProcArgs
		process.ProcArgs.List = []string{proc.CommandLine}

		// ProcExe - Cwd, Root not implemented
		process.ProcExe.Name = proc.ExecutablePath

		processes = append(processes, process)
	}
	self.List = processes
	return nil
}

func (self *ProcList) Get() error {
	var procs []Win32_Process
	whereClause := ""
	query := wmi.CreateQuery(&procs, whereClause)
	err := wmiClient.Query(query, &procs)
	if err != nil {
		return err
	}

	pids := make([]int, 0, len(procs))
	for _, proc := range procs {
		pids = append(pids, int(proc.ProcessId))
	}
	self.List = pids
	return nil
}

func getWmiWin32ProcessResult(pid int) (Win32_Process, error) {
	var procs []Win32_Process
	var proc Win32_Process

	query := wmi.CreateQuery(&procs, fmt.Sprintf("WHERE ProcessId = %d", pid))
	err := wmiClient.Query(query, &procs)
	if err != nil {
		return proc, err
	}

	if len(procs) == 0 {
		return proc, fmt.Errorf("Couldn't find pid %d", pid)
	}
	if len(procs) != 1 {
		// This shouldn't happen
		return proc, fmt.Errorf("Expected single WMI result")
	}
	return procs[0], nil
}

func (self *ProcState) Get(pid int) error {
	proc, err := getWmiWin32ProcessResult(pid)
	if err != nil {
		return err
	}

	self.Name = proc.Name
	self.Pid = int(proc.ProcessId)
	self.Ppid = int(proc.ParentProcessId)
	self.Priority = int(proc.Priority)
	self.State = convertWindowsRunState(WindowsRunState(proc.ExecutionState))

	return nil
}

func (self *ProcIo) Get(pid int) error {
	proc, err := getWmiWin32ProcessResult(pid)
	if err != nil {
		return err
	}

	self.ReadBytes = proc.ReadTransferCount
	self.ReadOps = proc.ReadOperationCount
	self.WriteBytes = proc.WriteTransferCount
	self.WriteOps = proc.WriteOperationCount
	return nil
}

func (self *ProcMem) Get(pid int) error {
	proc, err := getWmiWin32ProcessResult(pid)
	if err != nil {
		return err
	}

	self.Size = proc.VirtualSize
	self.Resident = proc.WorkingSetSize
	self.PageFaults = uint64(proc.PageFaults)

	// Share, MinorFaults and MajorFaults are not available from the Win32_Process WMI class
	return nil
}

func (self *ProcTime) Get(pid int) error {
	proc, err := getWmiWin32ProcessResult(pid)
	if err != nil {
		return err
	}

	self.User = convert100NsUnitsToMillis(proc.UserModeTime)
	self.Sys = convert100NsUnitsToMillis(proc.KernelModeTime)
	self.Total = self.User + self.Sys
	self.Total = self.User + self.Sys

	// Convert proc.CreationDate to millis
	self.StartTime = uint64(proc.CreationDate.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond)))
	return nil
}

func (self *ProcTime) CalculateCpuPercent(other *ProcTime) error {
	// CPU Percentage is already provided by Get()
	return nil
}

func (self *ProcArgs) Get(pid int) error {
	proc, err := getWmiWin32ProcessResult(pid)
	if err != nil {
		return err
	}

	self.List = []string{proc.CommandLine}
	return nil
}

func (self *ProcExe) Get(pid int) error {
	proc, err := getWmiWin32ProcessResult(pid)
	if err != nil {
		return err
	}

	self.Name = proc.ExecutablePath
	return nil
}

func (self *FileSystemUsage) Get(path string) error {
	var availableBytes C.ULARGE_INTEGER
	var totalBytes C.ULARGE_INTEGER
	var totalFreeBytes C.ULARGE_INTEGER

	pathChars := C.CString(path)
	defer C.free(unsafe.Pointer(pathChars))

	succeeded := C.GetDiskFreeSpaceEx((*C.CHAR)(pathChars), &availableBytes, &totalBytes, &totalFreeBytes)
	if succeeded == C.FALSE {
		lastError := C.GetLastError()
		return fmt.Errorf("GetDiskFreeSpaceEx failed with error: %d", int(lastError))
	}

	self.Total = *(*uint64)(unsafe.Pointer(&totalBytes))
	self.Avail = *(*uint64)(unsafe.Pointer(&availableBytes))
	self.Used = self.Total - self.Avail
	return nil
}

func (self *NetIfaceList) Get() error {
	netQueries := []string{
		`\Network Interface(*)\Bytes Sent/sec`,
		`\Network Interface(*)\Bytes Received/sec`,
		`\Network Interface(*)\Packets Sent/sec`,
		`\Network Interface(*)\Packets Received/sec`,
		`\Network Interface(*)\Packets outbound errors`,
		`\Network Interface(*)\Packets received errors`,
		`\Network Interface(*)\Packets outbound discarded`,
		`\Network Interface(*)\Packets received discarded`,
		`\Network Interface(*)\Packets received non-unicast/sec`,
	}
	queryResults, err := runRawPdhArrayQueries(netQueries)
	if err != nil {
		return err
	}
	self.List = make([]NetIface, 0)
	for iface, res := range queryResults {
		ifaceStruct := NetIface{
			Name:          iface,
			SendBytes:     res[0],
			RecvBytes:     res[1],
			SendPackets:   res[2],
			RecvPackets:   res[3],
			SendErrors:    res[4],
			RecvErrors:    res[5],
			SendDropped:   res[6],
			RecvDropped:   res[7],
			RecvMulticast: res[8],
		}
		self.List = append(self.List, ifaceStruct)
	}
	return nil
}

func convertTcpState(state C.DWORD) NetConnState {
	switch state {
	case C.MIB_TCP_STATE_CLOSED:
		return ConnStateClose
	case C.MIB_TCP_STATE_LISTEN:
		return ConnStateListen
	case C.MIB_TCP_STATE_SYN_SENT:
		return ConnStateSynSent
	case C.MIB_TCP_STATE_SYN_RCVD:
		return ConnStateSynRecv
	case C.MIB_TCP_STATE_ESTAB:
		return ConnStateEstablished
	case C.MIB_TCP_STATE_FIN_WAIT1:
		return ConnStateFinWait1
	case C.MIB_TCP_STATE_FIN_WAIT2:
		return ConnStateFinWait2
	case C.MIB_TCP_STATE_CLOSE_WAIT:
		return ConnStateCloseWait
	case C.MIB_TCP_STATE_CLOSING:
		return ConnStateClosing
	case C.MIB_TCP_STATE_LAST_ACK:
		return ConnStateLastAck
	case C.MIB_TCP_STATE_TIME_WAIT:
		return ConnStateTimeWait
	default:
		return 0
	}
}

/* Helper methods to convert IPv4 and IPv6 representations to Go byte arrays */
func UlongToBytes(addr C.u_long) []byte {
	return []byte{byte((addr & 0xFF000000) >> 24), byte((addr & 0x00FF0000) >> 16), byte((addr & 0x0000FF00) >> 8), byte(addr & 0x000000FF)}
}

func In6AddrToBytes(addr [16]C.UCHAR) []byte {
	outputAddr := make([]byte, 16)
	for i := 0; i < 16; i++ {
		outputAddr[i] = byte(addr[i])
	}
	return outputAddr
}

/* Helper methods to access rows of the MIB tables - Go doesn't know the size of the row arrays, we have to compute the offsets ourselves */
func tcpTableElement(table C.PMIB_TCPTABLE_OWNER_PID, index C.DWORD) C.PMIB_TCPROW_OWNER_PID {
	if index >= table.dwNumEntries {
		return nil
	}
	return C.PMIB_TCPROW_OWNER_PID(unsafe.Pointer(uintptr(unsafe.Pointer(&table.table)) + unsafe.Sizeof(table.table[0])*uintptr(index)))
}

func udpTableElement(table C.PMIB_UDPTABLE_OWNER_PID, index C.DWORD) C.PMIB_UDPROW_OWNER_PID {
	if index >= table.dwNumEntries {
		return nil
	}
	return C.PMIB_UDPROW_OWNER_PID(unsafe.Pointer(uintptr(unsafe.Pointer(&table.table)) + unsafe.Sizeof(table.table[0])*uintptr(index)))
}

func tcp6TableElement(table C.PMIB_TCP6TABLE_OWNER_PID, index C.DWORD) C.PMIB_TCP6ROW_OWNER_PID {
	if index >= table.dwNumEntries {
		return nil
	}
	return C.PMIB_TCP6ROW_OWNER_PID(unsafe.Pointer(uintptr(unsafe.Pointer(&table.table)) + unsafe.Sizeof(table.table[0])*uintptr(index)))
}

func udp6TableElement(table C.PMIB_UDP6TABLE_OWNER_PID, index C.DWORD) C.PMIB_UDP6ROW_OWNER_PID {
	if index >= table.dwNumEntries {
		return nil
	}
	return C.PMIB_UDP6ROW_OWNER_PID(unsafe.Pointer(uintptr(unsafe.Pointer(&table.table)) + unsafe.Sizeof(table.table[0])*uintptr(index)))
}

func getPidToProcessNameMap() (map[int]string, error) {
	// Query all processes
	var procs []Win32_Process
	whereClause := ""
	query := wmi.CreateQuery(&procs, whereClause)
	err := wmiClient.Query(query, &procs)
	if err != nil {
		return nil, err
	}

	pidMap := make(map[int]string)
	for _, proc := range procs {
		pidMap[int(proc.ProcessId)] = proc.Name
	}
	return pidMap, nil
}

func populateProcessName(netConnsPtr *[]NetConn) {
	pidMap, pidErr := getPidToProcessNameMap()
	if pidErr == nil {
		netConns := *netConnsPtr
		for i, _ := range netConns {
			netConns[i].ProcessName = pidMap[netConns[i].Pid]
		}
	}
}

func (self *NetTcpConnList) Get() error {
	var err C.DWORD
	table := C.getTcpTable(&err)
	if err != 0 {
		return fmt.Errorf("Error getting list of TCP connections: %v", err)
	}
	defer C.free(unsafe.Pointer(table))
	self.List = make([]NetConn, 0)
	for i := C.DWORD(0); i < table.dwNumEntries; i++ {
		elem := tcpTableElement(table, i)
		if elem == nil {
			return fmt.Errorf("Error getting connection %v, beyond array bounds", i)
		}
		localAddr := C.htonl(C.u_long(elem.dwLocalAddr))
		remoteAddr := C.htonl(C.u_long(elem.dwRemoteAddr))
		conn := NetConn{
			Proto:      ConnProtoTcp,
			Status:     convertTcpState(elem.dwState),
			LocalAddr:  UlongToBytes(localAddr),
			RemoteAddr: UlongToBytes(remoteAddr),
			LocalPort:  uint64(C.ntohs(C.u_short(elem.dwLocalPort))),
			RemotePort: uint64(C.ntohs(C.u_short(elem.dwRemotePort))),
			Pid:        int(elem.dwOwningPid),
		}
		self.List = append(self.List, conn)
	}

	populateProcessName(&self.List)
	return nil
}

func (self *NetUdpConnList) Get() error {
	var err C.DWORD
	table := C.getUdpTable(&err)
	if err != 0 {
		return fmt.Errorf("Error getting list of UDP connections: %v", err)
	}
	defer C.free(unsafe.Pointer(table))
	self.List = make([]NetConn, 0)
	for i := C.DWORD(0); i < table.dwNumEntries; i++ {
		elem := udpTableElement(table, i)
		if elem == nil {
			return fmt.Errorf("Error getting connection %v, beyond array bounds", i)
		}
		localAddr := C.htonl(C.u_long(elem.dwLocalAddr))
		conn := NetConn{
			LocalAddr: UlongToBytes(localAddr),
			LocalPort: uint64(C.ntohs(C.u_short(elem.dwLocalPort))),
			Proto:     ConnProtoUdp,
			Pid:       int(elem.dwOwningPid),
		}
		self.List = append(self.List, conn)
	}

	populateProcessName(&self.List)
	return nil
}

func (self *NetRawConnList) Get() error {
	return notImplemented()
}

func (self *NetTcpV6ConnList) Get() error {
	var err C.DWORD
	table := C.getTcp6Table(&err)
	if err != 0 {
		return fmt.Errorf("Error getting list of TCP connections: %v", err)
	}
	defer C.free(unsafe.Pointer(table))
	self.List = make([]NetConn, 0)
	for i := C.DWORD(0); i < table.dwNumEntries; i++ {
		elem := tcp6TableElement(table, i)
		if elem == nil {
			return fmt.Errorf("Error getting connection %v, beyond array bounds", i)
		}
		conn := NetConn{
			Proto:      ConnProtoTcp,
			Status:     convertTcpState(C.DWORD(elem.dwState)),
			LocalAddr:  In6AddrToBytes(elem.ucLocalAddr),
			RemoteAddr: In6AddrToBytes(elem.ucRemoteAddr),
			LocalPort:  uint64(C.ntohs(C.u_short(elem.dwLocalPort))),
			RemotePort: uint64(C.ntohs(C.u_short(elem.dwRemotePort))),
			Pid:        int(elem.dwOwningPid),
		}
		self.List = append(self.List, conn)
	}

	populateProcessName(&self.List)
	return nil
}

func (self *NetUdpV6ConnList) Get() error {
	var err C.DWORD
	table := C.getUdp6Table(&err)
	if err != 0 {
		return fmt.Errorf("Error getting list of UDP connections: %v", err)
	}
	defer C.free(unsafe.Pointer(table))
	self.List = make([]NetConn, 0)
	for i := C.DWORD(0); i < table.dwNumEntries; i++ {
		elem := udp6TableElement(table, i)
		if elem == nil {
			return fmt.Errorf("Error getting connection %v, beyond array bounds", i)
		}
		conn := NetConn{
			LocalAddr: In6AddrToBytes(elem.ucLocalAddr),
			LocalPort: uint64(C.ntohs(C.u_short(elem.dwLocalPort))),
			Proto:     ConnProtoUdp,
			Pid:       int(elem.dwOwningPid),
		}
		self.List = append(self.List, conn)
	}

	populateProcessName(&self.List)
	return nil
}

func (self *NetRawV6ConnList) Get() error {
	return notImplemented()
}

func (self *NetProtoV4Stats) Get() error {
	// List of PDH counters to gather. PDH counters are retreived "raw", meaning that per-second
	// counters are returned as monotonically increasing values despite their name
	protoV4Queries := []string{
		`\TCPv4\Connections Active`,
		`\TCPv4\Connections Passive`,
		`\TCPv4\Connection Failures`,
		`\TCPv4\Connections Reset`,
		`\TCPv4\Connections Established`,
		`\TCPv4\Segments Received/sec`,
		`\TCPv4\Segments Sent/sec`,
		`\TCPv4\Segments Retransmitted/sec`,

		`\UDPv4\Datagrams Received/sec`,
		`\UDPv4\Datagrams Sent/sec`,
		`\UDPv4\Datagrams Received Errors`,
		`\UDPv4\Datagrams No Port/sec`,

		`\IPv4\Datagrams Received/sec`,
		`\IPv4\Datagrams Received Header Errors`,
		`\IPv4\Datagrams Received Address Errors`,
		`\IPv4\Datagrams Forwarded/sec`,
		`\IPv4\Datagrams Received Delivered/sec`,
		`\IPv4\Datagrams Received Discarded`,
		`\IPv4\Datagrams Received Unknown Protocol`,
		`\IPv4\Datagrams Sent/sec`,
		`\IPv4\Datagrams Outbound Discarded`,
		`\IPv4\Datagrams Outbound No Route`,

		`\ICMP\Messages Received/sec`,
		`\ICMP\Messages Received Errors`,
		`\ICMP\Received Dest. Unreachable`,
		`\ICMP\Messages Sent/sec`,
		`\ICMP\Messages Outbound Errors`,
		`\ICMP\Sent Destination Unreachable`,
	}

	results, err := runRawPdhQueries(protoV4Queries)
	if err != nil {
		return err
	}
	if len(results) != len(protoV4Queries) {
		return errors.New("Incorrect results length")
	}

	self.TCP.ActiveOpens, results = uint64(results[0]), results[1:]
	self.TCP.PassiveOpens, results = uint64(results[0]), results[1:]
	self.TCP.AttemptFails, results = uint64(results[0]), results[1:]
	self.TCP.EstabResets, results = uint64(results[0]), results[1:]
	self.TCP.CurrEstab, results = uint64(results[0]), results[1:]
	self.TCP.InSegs, results = uint64(results[0]), results[1:]
	self.TCP.OutSegs, results = uint64(results[0]), results[1:]
	self.TCP.RetransSegs, results = uint64(results[0]), results[1:]
	// InErrs, OutRsts not available from PDH counters

	self.UDP.InDatagrams, results = uint64(results[0]), results[1:]
	self.UDP.OutDatagrams, results = uint64(results[0]), results[1:]
	self.UDP.InErrors, results = uint64(results[0]), results[1:]
	self.UDP.NoPorts, results = uint64(results[0]), results[1:]
	// RcvbufErrors, SndbufErrors not available from PDH counters

	self.IP.InReceives, results = uint64(results[0]), results[1:]
	self.IP.InHdrErrors, results = uint64(results[0]), results[1:]
	self.IP.InAddrErrors, results = uint64(results[0]), results[1:]
	self.IP.ForwDatagrams, results = uint64(results[0]), results[1:]
	self.IP.InDelivers, results = uint64(results[0]), results[1:]
	self.IP.InDiscards, results = uint64(results[0]), results[1:]
	self.IP.InUnknownProtos, results = uint64(results[0]), results[1:]
	self.IP.OutRequests, results = uint64(results[0]), results[1:]
	self.IP.OutDiscards, results = uint64(results[0]), results[1:]
	self.IP.OutNoRoutes, results = uint64(results[0]), results[1:]

	self.ICMP.InMsgs, results = uint64(results[0]), results[1:]
	self.ICMP.InErrors, results = uint64(results[0]), results[1:]
	self.ICMP.InDestUnreachs, results = uint64(results[0]), results[1:]
	self.ICMP.OutMsgs, results = uint64(results[0]), results[1:]
	self.ICMP.OutErrors, results = uint64(results[0]), results[1:]
	self.ICMP.OutDestUnreachs, results = uint64(results[0]), results[1:]

	return nil
}

func (self *NetProtoV6Stats) Get() error {
	// List of PDH counters to gather. PDH counters are retreived "raw", meaning that per-second
	// counters are returned as monotonically increasing values despite their name
	protoV6Queries := []string{
		`\TCPv6\Connections Active`,
		`\TCPv6\Connections Passive`,
		`\TCPv6\Connection Failures`,
		`\TCPv6\Connections Reset`,
		`\TCPv6\Connections Established`,
		`\TCPv6\Segments Received/sec`,
		`\TCPv6\Segments Sent/sec`,
		`\TCPv6\Segments Retransmitted/sec`,

		`\UDPv6\Datagrams Received/sec`,
		`\UDPv6\Datagrams Sent/sec`,
		`\UDPv6\Datagrams Received Errors`,
		`\UDPv6\Datagrams No Port/sec`,

		`\IPv6\Datagrams Received/sec`,
		`\IPv6\Datagrams Received Header Errors`,
		`\IPv6\Datagrams Received Address Errors`,
		`\IPv6\Datagrams Forwarded/sec`,
		`\IPv6\Datagrams Received Delivered/sec`,
		`\IPv6\Datagrams Received Discarded`,
		`\IPv6\Datagrams Received Unknown Protocol`,
		`\IPv6\Datagrams Sent/sec`,
		`\IPv6\Datagrams Outbound Discarded`,
		`\IPv6\Datagrams Outbound No Route`,

		`\ICMPv6\Messages Received/sec`,
		`\ICMPv6\Messages Received Errors`,
		`\ICMPv6\Received Dest. Unreachable`,
		`\ICMPv6\Messages Sent/sec`,
		`\ICMPv6\Messages Outbound Errors`,
		`\ICMPv6\Sent Destination Unreachable`,
	}

	results, err := runRawPdhQueries(protoV6Queries)
	if err != nil {
		return err
	}
	if len(results) != len(protoV6Queries) {
		return errors.New("Incorrect results length")
	}

	self.TCP.ActiveOpens, results = uint64(results[0]), results[1:]
	self.TCP.PassiveOpens, results = uint64(results[0]), results[1:]
	self.TCP.AttemptFails, results = uint64(results[0]), results[1:]
	self.TCP.EstabResets, results = uint64(results[0]), results[1:]
	self.TCP.CurrEstab, results = uint64(results[0]), results[1:]
	self.TCP.InSegs, results = uint64(results[0]), results[1:]
	self.TCP.OutSegs, results = uint64(results[0]), results[1:]
	self.TCP.RetransSegs, results = uint64(results[0]), results[1:]
	// InErrs, OutRsts not available from PDH counters

	self.UDP.InDatagrams, results = uint64(results[0]), results[1:]
	self.UDP.OutDatagrams, results = uint64(results[0]), results[1:]
	self.UDP.InErrors, results = uint64(results[0]), results[1:]
	self.UDP.NoPorts, results = uint64(results[0]), results[1:]
	// RcvbufErrors, SndbufErrors not available from PDH counters

	self.IP.InReceives, results = uint64(results[0]), results[1:]
	self.IP.InHdrErrors, results = uint64(results[0]), results[1:]
	self.IP.InAddrErrors, results = uint64(results[0]), results[1:]
	self.IP.ForwDatagrams, results = uint64(results[0]), results[1:]
	self.IP.InDelivers, results = uint64(results[0]), results[1:]
	self.IP.InDiscards, results = uint64(results[0]), results[1:]
	self.IP.InUnknownProtos, results = uint64(results[0]), results[1:]
	self.IP.OutRequests, results = uint64(results[0]), results[1:]
	self.IP.OutDiscards, results = uint64(results[0]), results[1:]
	self.IP.OutNoRoutes, results = uint64(results[0]), results[1:]

	self.ICMP.InMsgs, results = uint64(results[0]), results[1:]
	self.ICMP.InErrors, results = uint64(results[0]), results[1:]
	self.ICMP.InDestUnreachs, results = uint64(results[0]), results[1:]
	self.ICMP.OutMsgs, results = uint64(results[0]), results[1:]
	self.ICMP.OutErrors, results = uint64(results[0]), results[1:]
	self.ICMP.OutDestUnreachs, results = uint64(results[0]), results[1:]

	return nil
}

func (self *SystemInfo) Get() error {
	self.Sysname = "Windows"
	return nil
}

func (self *SystemDistribution) Get() error {
	self.Description = "Windows"
	return nil
}

func notImplemented() error {
	return ErrNotImplemented
}
