package sysinfo

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var processFile = "/proc/%d/%s"

type ProcessInfo struct {
	ID          uint64
	ParentID    uint64
	Name        string
	Path        string
	Arguments   []string
	State       string
	UserID      uint64
	GroupID     uint64
	GroupIDs    []uint64
	TTY         uint64
	ThreadCount uint64
	Priority    int64
	Nice        int64
	FDCount     uint64

	CPU    *ProcessCPUInfo
	Memory *ProcessMemoryInfo
}

type ProcessMemoryInfo struct {
	Virtual      uint64
	PeakVirtual  uint64
	Resident     uint64
	PeakResident uint64
	Locked       uint64
	Data         uint64
	Stack        uint64
	Text         uint64
	Shared       uint64
}

type ProcessCPUInfo struct {
	Start          uint64
	Total          uint64
	User           uint64
	System         uint64
	Guest          uint64
	ChildrenUser   uint64
	ChildrenSystem uint64
	ChildrenGuest  uint64

	sysUptime float64
}

func (pi *ProcessInfo) CPUUsagePercent() float64 {
	elapsed := pi.CPU.sysUptime - float64(pi.CPU.Start)/float64(TicksPerSecond)
	if elapsed == 0 {
		return 0.0
	}

	return 100 * float64(pi.CPU.Total) / float64(TicksPerSecond) / elapsed
}

func (pi *ProcessInfo) MemoryUsagePercent() (float64, error) {
	memory, err := Memory()
	if err != nil {
		return 0.0, err
	}

	return float64(pi.Memory.Resident) / float64(memory.Total) * 100, nil
}

func (pi *ProcessInfo) Update() error {
	if err := readProcStatusFile(pi.ID, pi); err != nil {
		return err
	}
	if err := readProcStatFile(pi.ID, pi); err != nil {
		return err
	}

	return nil
}

func ProcessCPUUsagePercent(a, b *ProcessCPUInfo) float64 {
	if a == nil || b == nil || (a.sysUptime == 0 && b.sysUptime == 0) {
		return 0.0
	}

	if a.sysUptime > b.sysUptime {
		a, b = b, a
	}

	elapsed := b.sysUptime - a.sysUptime
	if elapsed == 0 {
		elapsed = b.sysUptime
	}

	return 100 * float64(b.Total-a.Total) / float64(TicksPerSecond) / elapsed
}

func Process(pid uint64) (*ProcessInfo, error) {
	proc := &ProcessInfo{CPU: &ProcessCPUInfo{}, Memory: &ProcessMemoryInfo{}}
	if err := readProcStatusFile(pid, proc); err != nil {
		return nil, err
	}
	if err := readProcStatFile(pid, proc); err != nil {
		return nil, err
	}
	if err := readProcCmdlineFile(pid, proc); err != nil {
		return nil, err
	}

	return proc, nil
}

func ProcessList() ([]*ProcessInfo, error) {
	fis, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var procs []*ProcessInfo
	for _, fi := range fis {
		name := fi.Name()
		if n := name[0] - '0'; n < 0 || n > 9 {
			continue
		}

		pid, err := strconv.ParseUint(name, 0, 64)
		if err != nil {
			continue
		}

		proc, err := Process(pid)
		if err != nil {
			continue
		}

		procs = append(procs, proc)
	}

	return procs, nil
}

func readProcStatusFile(pid uint64, proc *ProcessInfo) error {
	file, err := os.Open(fmt.Sprintf(processFile, pid, "status"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		val := strings.ToLower(fields[1])
		key := strings.ToLower(strings.TrimSuffix(fields[0], ":"))
		switch key {
		case "name":
			proc.Name = strings.Trim(val, "()")

		case "tgid":
			proc.ID, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "ppid":
			proc.ParentID, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "state":
			proc.State = fields[1]

		case "threads":
			proc.ThreadCount, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "fdsize":
			proc.FDCount, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "uid":
			proc.UserID, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "gid":
			proc.GroupID, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "groups":
			for _, field := range fields[1:] {
				groupID, err := strconv.ParseUint(field, 10, 64)
				if err != nil {
					return err
				}

				proc.GroupIDs = append(proc.GroupIDs, groupID)
			}

		case "vmpeak":
			proc.Memory.PeakVirtual, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "vmsize":
			proc.Memory.Virtual, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "vmlck":
			proc.Memory.Locked, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "vmhwm":
			proc.Memory.PeakResident, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "vmrss":
			proc.Memory.Resident, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "vmdata":
			proc.Memory.Data, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "vmstk":
			proc.Memory.Stack, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "vmexe":
			proc.Memory.Text, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

		case "vmlib":
			proc.Memory.Shared, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

func readProcStatFile(pid uint64, proc *ProcessInfo) error {
	file, err := os.Open(fmt.Sprintf(processFile, pid, "stat"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 44 {
			return ErrInvalidFileFormat
		}

		proc.TTY, err = strconv.ParseUint(fields[6], 10, 64)
		if err != nil {
			return err
		}

		proc.CPU.User, err = strconv.ParseUint(fields[13], 10, 64)
		if err != nil {
			return err
		}

		proc.CPU.System, err = strconv.ParseUint(fields[14], 10, 64)
		if err != nil {
			return err
		}

		proc.CPU.ChildrenUser, err = strconv.ParseUint(fields[15], 10, 64)
		if err != nil {
			return err
		}

		proc.CPU.ChildrenSystem, err = strconv.ParseUint(fields[16], 10, 64)
		if err != nil {
			return err
		}

		proc.Priority, err = strconv.ParseInt(fields[17], 10, 64)
		if err != nil {
			return err
		}

		proc.Nice, err = strconv.ParseInt(fields[18], 10, 64)
		if err != nil {
			return err
		}

		proc.CPU.Start, err = strconv.ParseUint(fields[21], 10, 64)
		if err != nil {
			return err
		}

		proc.CPU.Guest, err = strconv.ParseUint(fields[42], 10, 64)
		if err != nil {
			return err
		}

		proc.CPU.ChildrenGuest, err = strconv.ParseUint(fields[43], 10, 64)
		if err != nil {
			return err
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	proc.CPU.User -= proc.CPU.Guest
	proc.CPU.ChildrenUser -= proc.CPU.ChildrenGuest
	proc.CPU.Total = proc.CPU.User + proc.CPU.System + proc.CPU.ChildrenUser +
		proc.CPU.ChildrenSystem + proc.CPU.Guest + proc.CPU.ChildrenGuest

	if proc.CPU.sysUptime, err = Uptime(); err != nil {
		return err
	}

	return nil
}

func readProcCmdlineFile(pid uint64, proc *ProcessInfo) error {
	cmd, err := readSingleValueFile(fmt.Sprintf(processFile, pid, "cmdline"))
	if err != nil {
		return nil
	}

	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return nil
	}

	proc.Path = fields[0]
	for _, arg := range fields[1:] {
		proc.Arguments = append(proc.Arguments, arg)
	}

	return nil
}
