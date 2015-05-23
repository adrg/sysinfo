package sysinfo

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const (
	cpuInfoPath    = "/proc/cpuinfo"
	cpuMinFreqPath = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_min_freq"
	cpuMaxFreqPath = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"
)

type CPU struct {
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

func CPUInfo() (*CPU, error) {
	var cpu = &CPU{}

	// read cpu max frequency
	minFreq, err := readSingleValueFile(cpuMinFreqPath)
	if err != nil {
		return nil, err
	}

	cpu.MinFreq, err = strconv.ParseUint(minFreq, 10, 64)
	if err != nil {
		return nil, err
	}

	// read cpu min frequency
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
