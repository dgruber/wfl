// Copyright (c) 2012 VMware, Inc.

package sigar

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var system struct {
	ticks uint64
	btime uint64
}

var Procd string
var Sysd string
var Etcd string

func init() {
	system.ticks = 100 // C.sysconf(C._SC_CLK_TCK)

	Procd = "/proc"
	Sysd = "/sys"
	Etcd = "/etc"
	LoadStartTime()
}

func LoadStartTime() {
	// grab system boot time
	readFile(Procd+"/stat", func(line string) bool {
		if strings.HasPrefix(line, "btime") {
			system.btime, _ = strtoull(line[6:])
			return false // stop reading
		}
		return true
	})
}

func (self *LoadAverage) Get() error {
	line, err := ioutil.ReadFile(Procd + "/loadavg")
	if err != nil {
		return nil
	}

	fields := strings.Fields(string(line))

	self.One, _ = strconv.ParseFloat(fields[0], 64)
	self.Five, _ = strconv.ParseFloat(fields[1], 64)
	self.Fifteen, _ = strconv.ParseFloat(fields[2], 64)

	return nil
}

func (self *Uptime) Get() error {
	sysinfo := syscall.Sysinfo_t{}

	if err := syscall.Sysinfo(&sysinfo); err != nil {
		return err
	}

	self.Length = float64(sysinfo.Uptime)

	return nil
}

func (self *Mem) Get() error {
	var buffers, cached uint64
	table := map[string]*uint64{
		"MemTotal": &self.Total,
		"MemFree":  &self.Free,
		"Buffers":  &buffers,
		"Cached":   &cached,
	}

	if err := parseMeminfo(table); err != nil {
		return err
	}

	self.Used = self.Total - self.Free
	kern := buffers + cached
	self.ActualFree = self.Free + kern
	self.ActualUsed = self.Used - kern

	return nil
}

func (self *Swap) Get() error {
	table := map[string]*uint64{
		"SwapTotal": &self.Total,
		"SwapFree":  &self.Free,
	}

	if err := parseMeminfo(table); err != nil {
		return err
	}

	self.Used = self.Total - self.Free
	return nil
}

func (self *Cpu) Get() error {
	return readFile(Procd+"/stat", func(line string) bool {
		if len(line) > 4 && line[0:4] == "cpu " {
			parseCpuStat(self, line)
			return false
		}
		return true

	})
}

func (self *CpuList) Get() error {
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 4
	}
	list := make([]Cpu, 0, capacity)

	err := readFile(Procd+"/stat", func(line string) bool {
		if len(line) > 3 && line[0:3] == "cpu" && line[3] != ' ' {
			cpu := Cpu{}
			parseCpuStat(&cpu, line)
			list = append(list, cpu)
		}
		return true
	})

	self.List = list

	return err
}
func (self *NetProtoV6Stats) Get() error {
	return readFile(Procd+"/net/snmp6", func(line string) bool {
		fields := strings.Fields(line)

		// Lines should be key/value pairs separated by whitespace, ignore other lines
		if len(fields) != 2 {
			return true
		}

		switch fields[0] {
		case "Ip6InReceives":
			self.IP.InReceives, _ = strtoull(fields[1])
		case "Ip6InAddrErrors":
			self.IP.InAddrErrors, _ = strtoull(fields[1])
		case "Ip6OutForwDatagrams":
			self.IP.ForwDatagrams, _ = strtoull(fields[1])
		case "Ip6InDelivers":
			self.IP.InDelivers, _ = strtoull(fields[1])
		case "Ip6InDiscards":
			self.IP.InDiscards, _ = strtoull(fields[1])
		case "Ip6OutRequests":
			self.IP.OutRequests, _ = strtoull(fields[1])
		case "Icmp6InMsgs":
			self.ICMP.InMsgs, _ = strtoull(fields[1])
		case "Icmp6InErrors":
			self.ICMP.InErrors, _ = strtoull(fields[1])
		case "Icmp6InDestUnreachs":
			self.ICMP.InDestUnreachs, _ = strtoull(fields[1])
		case "Icmp6OutMsgs":
			self.ICMP.OutMsgs, _ = strtoull(fields[1])
		case "Icmp6OutDestUnreachs":
			self.ICMP.OutDestUnreachs, _ = strtoull(fields[1])
		case "Udp6InDatagrams":
			self.UDP.InDatagrams, _ = strtoull(fields[1])
		case "Udp6OutDatagrams":
			self.UDP.OutDatagrams, _ = strtoull(fields[1])
		case "Udp6InErrors":
			self.UDP.InErrors, _ = strtoull(fields[1])
		case "Udp6NoPorts":
			self.UDP.NoPorts, _ = strtoull(fields[1])
		}
		return true
	})
}

func readField(positions map[string]int, fields []string, field string) uint64 {
	if positions[field] != 0 {
		value, _ := strtoull(fields[positions[field]])
		return value
	}
	return 0
}

func (self *NetProtoV4Stats) Get() error {
	// Each line starts with a header that describes the values, e.g.:
	// Udp: InDatagrams NoPorts InErrors OutDatagrams RcvbufErrors SndbufErrors
	// This map keeps track of the names of each position for each protocol. Reload
	// it each time we parse.
	protocols := make(map[string]map[string]int)

	return readFile(Procd+"/net/snmp", func(line string) bool {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return true
		}

		// The first field is the protocol name
		protocol := fields[0]

		// The first line with this protocol name defines the column names
		positions := protocols[protocol]

		// Parse column headers if needed
		if positions == nil {
			positions = make(map[string]int)
			for i := 1; i < len(fields); i++ {
				positions[fields[i]] = i
			}

			protocols[protocol] = positions
			return true
		}

		// Use the previously populated positions map to parse the line
		switch protocol {
		case "Ip:":
			self.IP.InReceives = readField(positions, fields, "InReceives")
			self.IP.InHdrErrors = readField(positions, fields, "InHdrErrors")
			self.IP.InAddrErrors = readField(positions, fields, "InAddrErrors")
			self.IP.ForwDatagrams = readField(positions, fields, "ForwDatagrams")
			self.IP.InDelivers = readField(positions, fields, "InDelivers")
			self.IP.InDiscards = readField(positions, fields, "InDiscards")
			self.IP.InUnknownProtos = readField(positions, fields, "InUnknownProtos")
			self.IP.OutRequests = readField(positions, fields, "OutRequests")
			self.IP.OutDiscards = readField(positions, fields, "OutDiscards")
			self.IP.OutNoRoutes = readField(positions, fields, "OutNoRoutes")

		case "Icmp:":
			self.ICMP.InMsgs = readField(positions, fields, "InMsgs")
			self.ICMP.InErrors = readField(positions, fields, "InErrors")
			self.ICMP.InDestUnreachs = readField(positions, fields, "InDestUnreachs")
			self.ICMP.OutMsgs = readField(positions, fields, "OutMsgs")
			self.ICMP.OutErrors = readField(positions, fields, "OutErrors")
			self.ICMP.OutDestUnreachs = readField(positions, fields, "OutDestUnreachs")

		case "Tcp:":
			self.TCP.ActiveOpens = readField(positions, fields, "ActiveOpens")
			self.TCP.PassiveOpens = readField(positions, fields, "PassiveOpens")
			self.TCP.AttemptFails = readField(positions, fields, "AttemptFails")
			self.TCP.EstabResets = readField(positions, fields, "EstabResets")
			self.TCP.CurrEstab = readField(positions, fields, "CurrEstab")
			self.TCP.InSegs = readField(positions, fields, "InSegs")
			self.TCP.OutSegs = readField(positions, fields, "OutSegs")
			self.TCP.RetransSegs = readField(positions, fields, "RetransSegs")
			self.TCP.InErrs = readField(positions, fields, "InErrs")
			self.TCP.OutRsts = readField(positions, fields, "OutRsts")

		case "Udp:":
			self.UDP.InDatagrams = readField(positions, fields, "InDatagrams")
			self.UDP.OutDatagrams = readField(positions, fields, "OutDatagrams")
			self.UDP.InErrors = readField(positions, fields, "InErrors")
			self.UDP.NoPorts = readField(positions, fields, "NoPorts")
			self.UDP.RcvbufErrors = readField(positions, fields, "RcvbufErrors")
			self.UDP.SndbufErrors = readField(positions, fields, "SndbufErrors")
		}
		return true
	})
}

func (self *NetIfaceList) Get() error {
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 10
	}
	ifaceList := make([]NetIface, 0, capacity)

	// Interface metrics come from `/proc/net/dev`
	err := readFile(Procd+"/net/dev", func(line string) bool {
		fields := strings.Fields(strings.TrimLeft(line, " \t"))
		if len(fields) == 0 {
			return true
		}

		// Interface names end with a colon, otherwise this is the header
		ifaceName := fields[0]
		if ifaceName[len(ifaceName)-1] != ':' {
			return true
		}

		if len(fields) != 17 {
			return true
		}

		iface := NetIface{}
		iface.Name = ifaceName[:len(ifaceName)-1]
		iface.SendBytes, _ = strtoull(fields[9])
		iface.RecvBytes, _ = strtoull(fields[1])
		iface.SendPackets, _ = strtoull(fields[10])
		iface.RecvPackets, _ = strtoull(fields[2])
		iface.SendCompressed, _ = strtoull(fields[16])
		iface.RecvCompressed, _ = strtoull(fields[7])
		iface.RecvMulticast, _ = strtoull(fields[8])

		iface.SendErrors, _ = strtoull(fields[11])
		iface.RecvErrors, _ = strtoull(fields[3])
		iface.SendDropped, _ = strtoull(fields[12])
		iface.RecvDropped, _ = strtoull(fields[4])
		iface.SendFifoErrors, _ = strtoull(fields[13])
		iface.RecvFifoErrors, _ = strtoull(fields[5])

		iface.RecvFramingErrors, _ = strtoull(fields[6])
		iface.SendCarrier, _ = strtoull(fields[15])
		iface.SendCollisions, _ = strtoull(fields[14])

		ifaceList = append(ifaceList, iface)

		return true
	})

	// Try to get MTU, MAC address and physical link status
	// This will only work on 2.6 kernels and above - see https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-class-net
	for i := range ifaceList {
		mtuFile := fmt.Sprintf("%v/class/net/%v/mtu", Sysd, ifaceList[i].Name)
		macFile := fmt.Sprintf("%v/class/net/%v/address", Sysd, ifaceList[i].Name)
		linkStatFile := fmt.Sprintf("%v/class/net/%v/carrier", Sysd, ifaceList[i].Name)

		ifaceList[i].MTU = ReadUint(readFileLine(mtuFile))
		ifaceList[i].Mac = readFileLine(macFile)

		linkStat := readFileLine(linkStatFile)
		switch linkStat {
		case "0":
			ifaceList[i].LinkStatus = "DOWN"
		case "1":
			ifaceList[i].LinkStatus = "UP"
		default:
			ifaceList[i].LinkStatus = "UNKNOWN"
		}
	}

	self.List = ifaceList
	return err
}

func (self *NetTcpConnList) Get() error {
	list, err := readConnList(Procd+"/net/tcp", 4, 17)
	if err != nil {
		return err
	}
	self.List = list
	return nil
}

func (self *NetUdpConnList) Get() error {
	list, err := readConnList(Procd+"/net/udp", 4, 13)
	if err != nil {
		return err
	}
	self.List = list
	return nil
}

func (self *NetRawConnList) Get() error {
	list, err := readConnList(Procd+"/net/raw", 4, 13)
	if err != nil {
		return err
	}
	self.List = list
	return nil
}

func (self *NetTcpV6ConnList) Get() error {
	list, err := readConnList(Procd+"/net/tcp6", 16, 17)
	if err != nil {
		return err
	}
	self.List = list
	return nil
}

func (self *NetUdpV6ConnList) Get() error {
	list, err := readConnList(Procd+"/net/udp6", 16, 13)
	if err != nil {
		return err
	}
	self.List = list
	return nil
}

func (self *NetRawV6ConnList) Get() error {
	list, err := readConnList(Procd+"/net/raw6", 16, 13)
	if err != nil {
		return err
	}
	self.List = list
	return nil
}

/* Reads the format of the /proc/net/<proto> files, which have 2 header lines and a
   list of open connections. Different protocols have different numbers of trailing fields,
   but the first 5 are the same. */
func readConnList(listFile string, ipSizeBytes, numFields int) ([]NetConn, error) {
	connList := make([]NetConn, 0)
	err := readFile(listFile, func(line string) bool {
		fields := strings.Fields(line)
		if len(fields) != numFields {
			return true
		}
		// Skip the header, only take lines where the first field is <number>:
		if fields[0][len(fields[0])-1] != ':' {
			return true
		}

		var err error
		var conn NetConn
		conn.LocalAddr, conn.LocalPort, err = readConnIp(fields[1], ipSizeBytes)
		if err != nil {
			return true
		}

		conn.RemoteAddr, conn.RemotePort, err = readConnIp(fields[2], ipSizeBytes)
		if err != nil {
			return true
		}

		status, err := strconv.ParseInt(fields[3], 16, 8)
		if err != nil {
			return true
		}

		conn.Status = NetConnState(status)
		queues := strings.Split(fields[4], ":")
		if len(queues) != 2 {
			return true
		}

		conn.SendQueue, err = strtoull(queues[0])
		if err != nil {
			return true
		}

		conn.RecvQueue, err = strtoull(queues[1])
		if err != nil {
			return true
		}

		connList = append(connList, conn)
		return true
	})
	return connList, err
}

/* Decode an IP:port pair, with either a 16 or 4-byte address and 2-byte port,
   both hex-encoded. TODO: Test on a big-endian architecture. */
func readConnIp(field string, lenBytes int) (net.IP, uint64, error) {
	parts := strings.Split(field, ":")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("Unable to split into IP and port")
	}
	if len(parts[0]) != lenBytes*2 {
		return nil, 0, fmt.Errorf("Unable to parse IP, expected %v bytes got %v", lenBytes, len(parts[0]))
	}
	var port int64
	var err error

	port, err = strconv.ParseInt(parts[1], 16, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to parse port, %v - %v", parts[1], err)
	}

	ip := make([]byte, lenBytes)
	// The 32-bit words are in order, but the words themselves are little-endian
	for i := 0; i < lenBytes; i += 4 {
		for j := 0; j < 4; j++ {
			byteVal, err := strconv.ParseInt(parts[0][(j+i)*2:(j+i+1)*2], 16, 8)
			if err != nil {
				return nil, 0, fmt.Errorf("Unable to parse IP, %v - %v", parts[0], err)
			}
			ip[i+(3-j)] = byte(byteVal)
		}
	}
	return net.IP(ip), uint64(port), nil
}

func (self *FileSystemList) Get() error {
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 10
	}
	fslist := make([]FileSystem, 0, capacity)

	err := readFile("/etc/mtab", func(line string) bool {
		fields := strings.Fields(line)

		fs := FileSystem{}
		fs.DevName = fields[0]
		fs.DirName = fields[1]
		fs.SysTypeName = fields[2]
		fs.Options = fields[3]

		fslist = append(fslist, fs)

		return true
	})

	self.List = fslist

	return err
}

func (self *DiskList) Get() error {
	/* List all the partitions, and check the major/minor device ID
	   to find which are devices vs. partitions (ex. sda v. sda1) */
	devices := make(map[string]bool)
	diskList := make(map[string]DiskIo)
	err := readFile(Procd+"/partitions", func(line string) bool {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			return true
		}
		majorDevId, err := strtoull(fields[0])
		if err != nil {
			return true
		}
		minorDevId, err := strtoull(fields[1])
		if err != nil {
			return true
		}
		if isNotPartition(majorDevId, minorDevId) {
			devices[fields[3]] = true
		}
		return true
	})

	/* Get all device stats from /proc/diskstats and filter by
	   devices from /proc/partitions */
	err = readFile(Procd+"/diskstats", func(line string) bool {
		fields := strings.Fields(line)
		if len(fields) < 13 {
			return true
		}
		deviceName := fields[2]
		if _, ok := devices[deviceName]; !ok {
			return true
		}
		io := DiskIo{}
		io.ReadOps, _ = strtoull(fields[3])
		readBytes, _ := strtoull(fields[5])
		io.ReadBytes = readBytes * 512
		io.ReadTimeMs, _ = strtoull(fields[6])
		io.WriteOps, _ = strtoull(fields[7])
		writeBytes, _ := strtoull(fields[9])
		io.WriteBytes = writeBytes * 512
		io.WriteTimeMs, _ = strtoull(fields[10])
		io.IoTimeMs, _ = strtoull(fields[12])
		diskList[deviceName] = io
		return true
	})
	self.List = diskList
	return err
}

func (self *ProcList) Get() error {
	dir, err := os.Open(Procd)
	if err != nil {
		return err
	}
	defer dir.Close()

	const readAllDirnames = -1 // see os.File.Readdirnames doc

	names, err := dir.Readdirnames(readAllDirnames)
	if err != nil {
		return err
	}

	capacity := len(names)
	list := make([]int, 0, capacity)

	for _, name := range names {
		if name[0] < '0' || name[0] > '9' {
			continue
		}
		pid, err := strconv.Atoi(name)
		if err == nil {
			list = append(list, pid)
		}
	}

	self.List = list

	return nil
}

func (self *ProcIo) Get(pid int) error {
	assignMap := map[string]*uint64{
		"syscr:":       &self.ReadOps,
		"syscw:":       &self.WriteOps,
		"read_bytes:":  &self.ReadBytes,
		"write_bytes:": &self.WriteBytes,
	}
	err := readFile(fmt.Sprintf("%v/%v/io", Procd, pid), func(line string) bool {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return true
		}
		val, ok := assignMap[fields[0]]
		if !ok {
			return true
		}
		*val, _ = strtoull(fields[1])
		return true
	})
	return err
}

func (self *ProcState) Get(pid int) error {
	contents, err := readProcFile(pid, "stat")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(contents))

	self.Name = fields[1][1 : len(fields[1])-1] // strip ()'s

	self.State = RunState(fields[2][0])

	self.Ppid, _ = strconv.Atoi(fields[3])

	self.Tty, _ = strconv.Atoi(fields[6])

	self.Priority, _ = strconv.Atoi(fields[17])

	self.Nice, _ = strconv.Atoi(fields[18])

	self.Processor, _ = strconv.Atoi(fields[38])

	return nil
}

func (self *ProcMem) Get(pid int) error {
	contents, err := readProcFile(pid, "statm")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(contents))

	size, _ := strtoull(fields[0])
	self.Size = size << 12

	rss, _ := strtoull(fields[1])
	self.Resident = rss << 12

	share, _ := strtoull(fields[2])
	self.Share = share << 12

	contents, err = readProcFile(pid, "stat")
	if err != nil {
		return err
	}

	fields = strings.Fields(string(contents))

	self.MinorFaults, _ = strtoull(fields[10])
	self.MajorFaults, _ = strtoull(fields[12])
	self.PageFaults = self.MinorFaults + self.MajorFaults

	return nil
}

func (self *ProcTime) Get(pid int) error {
	contents, err := readProcFile(pid, "stat")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(contents))

	user, _ := strtoull(fields[13])
	sys, _ := strtoull(fields[14])
	// convert to millis
	self.User = user * (1000 / system.ticks)
	self.Sys = sys * (1000 / system.ticks)
	self.Total = self.User + self.Sys

	// convert to millis
	self.StartTime, _ = strtoull(fields[21])
	self.StartTime /= system.ticks
	self.StartTime += system.btime
	self.StartTime *= 1000

	return nil
}

func (self *ProcArgs) Get(pid int) error {
	contents, err := readProcFile(pid, "cmdline")
	if err != nil {
		return err
	}

	bbuf := bytes.NewBuffer(contents)

	var args []string

	for {
		arg, err := bbuf.ReadBytes(0)
		if err == io.EOF {
			break
		}
		args = append(args, string(chop(arg)))
	}

	self.List = args

	return nil
}

func (self *ProcExe) Get(pid int) error {
	fields := map[string]*string{
		"exe":  &self.Name,
		"cwd":  &self.Cwd,
		"root": &self.Root,
	}

	for name, field := range fields {
		val, err := os.Readlink(procFileName(pid, name))

		if err != nil {
			return err
		}

		*field = val
	}

	return nil
}

func parseMeminfo(table map[string]*uint64) error {
	return readFile(Procd+"/meminfo", func(line string) bool {
		fields := strings.Split(line, ":")

		if ptr := table[fields[0]]; ptr != nil {
			num := strings.TrimLeft(fields[1], " ")
			val, err := strtoull(strings.Fields(num)[0])
			if err == nil {
				*ptr = val * 1024
			}
		}

		return true
	})
}

func parseCpuStat(self *Cpu, line string) error {
	fields := strings.Fields(line)

	self.User, _ = strtoull(fields[1])
	self.Nice, _ = strtoull(fields[2])
	self.Sys, _ = strtoull(fields[3])
	self.Idle, _ = strtoull(fields[4])
	self.Wait, _ = strtoull(fields[5])
	self.Irq, _ = strtoull(fields[6])
	self.SoftIrq, _ = strtoull(fields[7])
	self.Stolen, _ = strtoull(fields[8])
	/* Guest was added in 2.6, not available on all kernels */
	if len(fields) > 9 {
		self.Guest, _ = strtoull(fields[9])
	}

	return nil
}

func readFile(file string, handler func(string) bool) error {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(bytes.NewBuffer(contents))

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if !handler(string(line)) {
			break
		}
	}

	return nil
}

// Read the first line of a file, ignoring any error
func readFileLine(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// Convert the value to an uint64, ignoring any error
func ReadUint(val string) uint64 {
	result, err := strconv.ParseUint(val, 10, 64)
	if err == nil {
		return result
	}
	return 0
}

func strtoull(val string) (uint64, error) {
	return strconv.ParseUint(val, 10, 64)
}

func procFileName(pid int, name string) string {
	return Procd + "/" + strconv.Itoa(pid) + "/" + name
}

func readProcFile(pid int, name string) ([]byte, error) {
	path := procFileName(pid, name)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		if perr, ok := err.(*os.PathError); ok {
			if perr.Err == syscall.ENOENT {
				return nil, syscall.ESRCH
			}
		}
	}

	return contents, err
}

/* For SCSI and IDE devices, only display devices and not individual partitions.
   For other major numbers, show all devices regardless of minor (for LVM, for example).
   As described here: http://www.linux-tutorial.info/modules.php?name=MContent&pageid=94 */
func isNotPartition(majorDevId, minorDevId uint64) bool {
	if majorDevId == 3 || majorDevId == 22 || // IDE0_MAJOR IDE1_MAJOR
		majorDevId == 33 || majorDevId == 34 || // IDE2_MAJOR IDE3_MAJOR
		majorDevId == 56 || majorDevId == 57 || // IDE4_MAJOR IDE5_MAJOR
		(majorDevId >= 88 && majorDevId <= 91) { // IDE6_MAJOR to IDE_IDE9_MAJOR
		return (minorDevId & 0x3F) == 0 // IDE uses bottom 10 bits for partitions
	}
	if majorDevId == 8 || // SCSI_DISK0_MAJOR
		(majorDevId >= 65 && majorDevId <= 71) || // SCSI_DISK1_MAJOR to SCSI_DISK7_MAJOR
		(majorDevId >= 128 && majorDevId <= 135) { // SCSI_DISK8_MAJOR to SCSI_DISK15_MAJOR
		return (minorDevId & 0x0F) == 0 // SCSI uses bottom 8 bits for partitions
	}
	return true
}

func (self *SystemInfo) Get() error {
	var uname syscall.Utsname
	err := syscall.Uname(&uname)
	if err != nil {
		return err
	}
	self.Sysname = bytePtrToString(&uname.Sysname[0])
	self.Nodename = bytePtrToString(&uname.Nodename[0])
	self.Release = bytePtrToString(&uname.Release[0])
	self.Version = bytePtrToString(&uname.Version[0])
	self.Machine = bytePtrToString(&uname.Machine[0])
	self.Domainname = bytePtrToString(&uname.Domainname[0])

	return nil
}

var distributionDesc string = "DISTRIB_DESCRIPTION="

func (self *SystemDistribution) Get() error {
	// Special case for redhat/centos, ignoring any error
	_ = readFile(Etcd+"/redhat-release", func(line string) bool {
		self.Description = line
		return false
	})
	if self.Description != "" {
		return nil
	}

	// Read /etc/lsb-release
	return readFile(Etcd+"/lsb-release", func(line string) bool {
		if strings.HasPrefix(line, distributionDesc) {
			self.Description = strings.Trim(line[len(distributionDesc):], `"`)
			return false
		}
		return true
	})
}
