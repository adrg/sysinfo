package sysinfo

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const (
	cpuStatPath    = "/proc/stat"
	cpuInfoPath    = "/proc/cpuinfo"
	cpuMinFreqPath = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_min_freq"
	cpuMaxFreqPath = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"
)

type CPUInfo struct {
	Name        string
	Model       uint64
	Family      uint64
	VendorID    string
	Stepping    string
	Cache       uint64
	Flags       []string
	MinFreq     uint64
	MaxFreq     uint64
	CoreCount   uint64
	ThreadCount uint64
	SocketCount uint64
}

type CPUStat struct {
	Total     uint64
	User      uint64
	Nice      uint64
	System    uint64
	Idle      uint64
	IOWait    uint64
	IRQ       uint64
	SoftIRQ   uint64
	Steal     uint64
	Guest     uint64
	GuestNice uint64
}

func CPU() (*CPUInfo, error) {
	var cpu = &CPUInfo{}

	// read CPU max frequency
	minFreq, err := readSingleValueFile(cpuMinFreqPath)
	if err != nil {
		return nil, err
	}

	cpu.MinFreq, err = strconv.ParseUint(minFreq, 10, 64)
	if err != nil {
		return nil, err
	}

	// read CPU min frequency
	maxFreq, err := readSingleValueFile(cpuMaxFreqPath)
	if err != nil {
		return nil, err
	}

	cpu.MaxFreq, err = strconv.ParseUint(maxFreq, 10, 64)
	if err != nil {
		return nil, err
	}

	// parse cpuinfo file
	file, err := os.Open(cpuInfoPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cores := map[uint64]struct{}{}
	sockets := map[uint64]struct{}{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			return nil, ErrInvalidFileFormat
		}

		value := strings.TrimSpace(fields[1])
		key := strings.ToLower(strings.TrimSpace(fields[0]))
		switch key {
		case "processor":
			cpu.ThreadCount++

		case "model name":
			if cpu.Name == "" {
				cpu.Name = value
			}

		case "model":
			if cpu.Model != 0 {
				continue
			}

			cpu.Model, err = strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}

		case "cpu family":
			if cpu.Family != 0 {
				continue
			}

			cpu.Family, err = strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}

		case "vendor_id":
			if cpu.VendorID == "" {
				cpu.VendorID = value
			}

		case "stepping":
			if cpu.Stepping == "" {
				cpu.Stepping = value
			}

		case "cache size":
			if cpu.Cache != 0 {
				continue
			}

			cacheFields := strings.Fields(value)
			if len(cacheFields) != 2 {
				return nil, ErrInvalidFileFormat
			}

			cpu.Cache, err = strconv.ParseUint(cacheFields[0], 10, 64)
			if err != nil {
				return nil, err
			}

		case "physical id":
			socket, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}

			if _, ok := sockets[socket]; !ok {
				sockets[socket] = struct{}{}
			}

		case "core id":
			core, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}

			if _, ok := cores[core]; !ok {
				cores[core] = struct{}{}
			}

		case "flags":
			if len(cpu.Flags) > 0 {
				continue
			}

			cpu.Flags = strings.Fields(value)
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	cpu.CoreCount = uint64(len(cores))
	cpu.SocketCount = uint64(len(sockets))

	return cpu, nil
}

func CPUUsage() (usageGlobal *CPUStat, usagePerCore []*CPUStat, err error) {
	file, err := os.Open(cpuStatPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "cpu") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 11 {
			return nil, nil, err
		}

		usage := &CPUStat{}
		usage.User, err = strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.Nice, err = strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.System, err = strconv.ParseUint(fields[3], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.Idle, err = strconv.ParseUint(fields[4], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.IOWait, err = strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.IRQ, err = strconv.ParseUint(fields[6], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.SoftIRQ, err = strconv.ParseUint(fields[7], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.Steal, err = strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.Guest, err = strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		usage.GuestNice, err = strconv.ParseUint(fields[10], 10, 64)
		if err != nil {
			return nil, nil, err
		}

		usage.User -= usage.Guest
		usage.Nice -= usage.GuestNice
		usage.Total = usage.User + usage.Nice + usage.System + usage.Idle +
			usage.IOWait + usage.IRQ + usage.SoftIRQ + usage.Steal +
			usage.Guest + usage.GuestNice

		if fields[0] == "cpu" {
			usageGlobal = usage
			continue
		}

		usagePerCore = append(usagePerCore, usage)
	}

	if err = scanner.Err(); err != nil {
		return nil, nil, err
	}

	return
}

func CPULoadPercent(firstSample *CPUStat, secondSample *CPUStat) float64 {
	if firstSample == nil || secondSample == nil {
		return 0.0
	}
	if firstSample.Total == 0 && secondSample.Total == 0 {
		return 0.0
	}

	if firstSample.Total > secondSample.Total {
		firstSample, secondSample = secondSample, firstSample
	}

	deltaTotal := secondSample.Total - firstSample.Total
	if deltaTotal == 0 {
		deltaTotal = secondSample.Total
	}

	deltaIdle := (secondSample.Idle + secondSample.IOWait) -
		(firstSample.Idle + firstSample.IOWait)
	if deltaIdle == 0 {
		deltaIdle = secondSample.Idle
	}

	return float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
}
