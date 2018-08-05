// Copyright (c) 2012 VMware, Inc.

package sigar

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
PMIB_TCPTABLE getTcpTable(PDWORD err) {
	PMIB_TCPTABLE pTable = (MIB_TCPTABLE *) malloc(sizeof(MIB_TCPTABLE));
	if (pTable == NULL) {
		*err = 1;
		return NULL;
	}
	DWORD size = sizeof(MIB_TCPTABLE);
	if ((*err = GetTcpTable(pTable, &size, FALSE)) != ERROR_INSUFFICIENT_BUFFER) {
		if (*err == NO_ERROR) {
			return pTable;
		}
		free(pTable);
		return NULL;
	}
	free(pTable);
	pTable = (MIB_TCPTABLE *) malloc(size);
	if ((*err = GetTcpTable(pTable, &size, FALSE)) != NO_ERROR) {
		free(pTable);
		return NULL;
	}
	*err = 0;
	return pTable;
}

PMIB_UDPTABLE getUdpTable(PDWORD err) {
	PMIB_UDPTABLE pTable = (MIB_UDPTABLE *) malloc(sizeof(MIB_UDPTABLE));
	if (pTable == NULL) {
		*err = 1;
		return NULL;
	}
	DWORD size = sizeof(MIB_UDPTABLE);
	if ((*err = GetUdpTable(pTable, &size, FALSE)) != ERROR_INSUFFICIENT_BUFFER) {
		if (*err == NO_ERROR) {
			return pTable;
		}
		free(pTable);
		return NULL;
	}
	free(pTable);
	pTable = (MIB_UDPTABLE *) malloc(size);
	if ((*err = GetUdpTable(pTable, &size, FALSE)) != NO_ERROR) {
		free(pTable);
		return NULL;
	}
	*err = 0;
	return pTable;
}

PMIB_TCP6TABLE getTcp6Table(PDWORD err) {
	PMIB_TCP6TABLE pTable = (MIB_TCP6TABLE *) malloc(sizeof(MIB_TCP6TABLE));
	if (pTable == NULL) {
		*err = 1;
		return NULL;
	}
	DWORD size = sizeof(MIB_TCP6TABLE);
	if ((*err = GetTcp6Table(pTable, &size, FALSE)) != ERROR_INSUFFICIENT_BUFFER) {
		if (*err == NO_ERROR) {
			return pTable;
		}
		free(pTable);
		return NULL;
	}
	free(pTable);
	pTable = (MIB_TCP6TABLE *) malloc(size);
	if ((*err = GetTcp6Table(pTable, &size, FALSE)) != NO_ERROR) {
		free(pTable);
		return NULL;
	}
	*err = 0;
	return pTable;
}

PMIB_UDP6TABLE getUdp6Table(PDWORD err) {
	PMIB_UDP6TABLE pTable = (MIB_UDP6TABLE *) malloc(sizeof(MIB_UDP6TABLE));
	if (pTable == NULL) {
		*err = 1;
		return NULL;
	}
	DWORD size = sizeof(MIB_UDP6TABLE);
	if ((*err = GetUdp6Table(pTable, &size, FALSE)) != ERROR_INSUFFICIENT_BUFFER) {
		if (*err == NO_ERROR) {
			return pTable;
		}
		free(pTable);
		return NULL;
	}
	free(pTable);
	pTable = (MIB_UDP6TABLE *) malloc(size);
	if ((*err = GetUdp6Table(pTable, &size, FALSE)) != NO_ERROR) {
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

func init() {
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

func (self *ProcList) Get() error {
	return notImplemented()
}

func (self *ProcState) Get(pid int) error {
	return notImplemented()
}

func (self *ProcMem) Get(pid int) error {
	return notImplemented()
}

func (self *ProcTime) Get(pid int) error {
	return notImplemented()
}

func (self *ProcArgs) Get(pid int) error {
	return notImplemented()
}

func (self *ProcExe) Get(pid int) error {
	return notImplemented()
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

func In6AddrToBytes(addr C.IN6_ADDR) []byte {
	outputAddr := make([]byte, 16)
	for i := 0; i < 16; i++ {
		outputAddr[i] = byte(addr.u[i])
	}
	return outputAddr
}

/* Helper methods to access rows of the MIB tables - Go doesn't know the size of the row arrays, we have to compute the offsets ourselves */
func tcpTableElement(table C.PMIB_TCPTABLE, index C.DWORD) C.PMIB_TCPROW {
	if index >= table.dwNumEntries {
		return nil
	}
	return C.PMIB_TCPROW(unsafe.Pointer(uintptr(unsafe.Pointer(&table.table)) + unsafe.Sizeof(table.table[0])*uintptr(index)))
}

func udpTableElement(table C.PMIB_UDPTABLE, index C.DWORD) C.PMIB_UDPROW {
	if index >= table.dwNumEntries {
		return nil
	}
	return C.PMIB_UDPROW(unsafe.Pointer(uintptr(unsafe.Pointer(&table.table)) + unsafe.Sizeof(table.table[0])*uintptr(index)))
}

func tcp6TableElement(table C.PMIB_TCP6TABLE, index C.DWORD) C.PMIB_TCP6ROW {
	if index >= table.dwNumEntries {
		return nil
	}
	return C.PMIB_TCP6ROW(unsafe.Pointer(uintptr(unsafe.Pointer(&table.table)) + unsafe.Sizeof(table.table[0])*uintptr(index)))
}

func udp6TableElement(table C.PMIB_UDP6TABLE, index C.DWORD) C.PMIB_UDP6ROW {
	if index >= table.dwNumEntries {
		return nil
	}
	return C.PMIB_UDP6ROW(unsafe.Pointer(uintptr(unsafe.Pointer(&table.table)) + unsafe.Sizeof(table.table[0])*uintptr(index)))
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
			Status:     convertTcpState(elem.dwState),
			LocalAddr:  UlongToBytes(localAddr),
			RemoteAddr: UlongToBytes(remoteAddr),
			LocalPort:  uint64(C.ntohs(C.u_short(elem.dwLocalPort))),
			RemotePort: uint64(C.ntohs(C.u_short(elem.dwRemotePort))),
		}
		self.List = append(self.List, conn)
	}
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
		}
		self.List = append(self.List, conn)
	}
	return nil
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
			Status:     convertTcpState(C.DWORD(elem.State)),
			LocalAddr:  In6AddrToBytes(elem.LocalAddr),
			RemoteAddr: In6AddrToBytes(elem.RemoteAddr),
			LocalPort:  uint64(C.ntohs(C.u_short(elem.dwLocalPort))),
			RemotePort: uint64(C.ntohs(C.u_short(elem.dwRemotePort))),
		}
		self.List = append(self.List, conn)
	}
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
			LocalAddr: In6AddrToBytes(elem.dwLocalAddr),
			LocalPort: uint64(C.ntohs(C.u_short(elem.dwLocalPort))),
		}
		self.List = append(self.List, conn)
	}
	return nil
}

func (self *NetProtoV4Stats) Get() error {
	protoV4Queries := []string{
		`\TCPv4\Connections Active`,
		`\TCPv4\Connections Passive`,
		`\TCPv4\Connection Failures`,
		`\TCPv4\Connections Reset`,
		`\TCPv4\Connections Established`,

		`\UDPv4\Datagrams Received Errors`,

		`\IPv4\Datagrams Received Header Errors`,
		`\IPv4\Datagrams Received Address Errors`,
		`\IPv4\Datagrams Received Discarded`,
		`\IPv4\Datagrams Received Unknown Protocol`,
		`\IPv4\Datagrams Outbound Discarded`,
		`\IPv4\Datagrams Outbound No Route`,

		`\ICMP\Messages Received Errors`,
		`\ICMP\Received Dest. Unreachable`,
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

	self.UDP.InErrors, results = uint64(results[0]), results[1:]

	self.IP.InHdrErrors, results = uint64(results[0]), results[1:]
	self.IP.InAddrErrors, results = uint64(results[0]), results[1:]
	self.IP.InDiscards, results = uint64(results[0]), results[1:]
	self.IP.InUnknownProtos, results = uint64(results[0]), results[1:]
	self.IP.OutDiscards, results = uint64(results[0]), results[1:]
	self.IP.OutNoRoutes, results = uint64(results[0]), results[1:]

	self.ICMP.InErrors, results = uint64(results[0]), results[1:]
	self.ICMP.InDestUnreachs, results = uint64(results[0]), results[1:]
	self.ICMP.OutErrors, results = uint64(results[0]), results[1:]
	self.ICMP.OutDestUnreachs, results = uint64(results[0]), results[1:]

	return nil
}

func (self *NetProtoV6Stats) Get() error {
	protoV6Queries := []string{
		`\TCPv6\Connections Active`,
		`\TCPv6\Connections Passive`,
		`\TCPv6\Connection Failures`,
		`\TCPv6\Connections Reset`,
		`\TCPv6\Connections Established`,

		`\UDPv6\Datagrams Received Errors`,

		`\IPv6\Datagrams Received Header Errors`,
		`\IPv6\Datagrams Received Address Errors`,
		`\IPv6\Datagrams Received Discarded`,
		`\IPv6\Datagrams Received Unknown Protocol`,
		`\IPv6\Datagrams Outbound Discarded`,
		`\IPv6\Datagrams Outbound No Route`,

		`\ICMPv6\Messages Received Errors`,
		`\ICMPv6\Received Dest. Unreachable`,
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

	self.UDP.InErrors, results = uint64(results[0]), results[1:]

	self.IP.InHdrErrors, results = uint64(results[0]), results[1:]
	self.IP.InAddrErrors, results = uint64(results[0]), results[1:]
	self.IP.InDiscards, results = uint64(results[0]), results[1:]
	self.IP.InUnknownProtos, results = uint64(results[0]), results[1:]
	self.IP.OutDiscards, results = uint64(results[0]), results[1:]
	self.IP.OutNoRoutes, results = uint64(results[0]), results[1:]

	self.ICMP.InErrors, results = uint64(results[0]), results[1:]
	self.ICMP.InDestUnreachs, results = uint64(results[0]), results[1:]
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
