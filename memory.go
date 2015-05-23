package sysinfo

import (
	"bufio"

	"os"
	"strconv"
	"strings"

	"github.com/adrg/unit"
)

type Memory struct {
	Total      uint64
	Free       uint64
	Used       uint64
	Cached     uint64
	Active     uint64
	Inactive   uint64
	SwapTotal  uint64
	SwapFree   uint64
	SwapUsed   uint64
	SwapCached uint64
	Buffers    uint64

	Unit unit.Memory
}

func (m *Memory) PercentUsed() float64 {
	return float64(m.Used) / float64(m.Total) * 100.0
}

func (m *Memory) PercentFree() float64 {
	return float64(m.Free) / float64(m.Total) * 100.0
}

func (m *Memory) PercentSwapUsed() float64 {
	return float64(m.SwapUsed) / float64(m.SwapTotal) * 100.0
}

func (m *Memory) PercentSwapFree() float64 {
	return float64(m.SwapFree) / float64(m.SwapTotal) * 100.0
}

func MemoryInfo() (*Memory, error) {
	mem := &Memory{Unit: unit.Kibibyte}

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			return nil, ErrInvalidFileFormat
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return nil, ErrInvalidFileFormat
		}

		section := strings.ToLower(strings.TrimSuffix(fields[0], ":"))
		switch section {
		case "memtotal":
			mem.Total = value
		case "memfree":
			mem.Free = value
		case "cached":
			mem.Cached = value
		case "swaptotal":
			mem.SwapTotal = value
		case "swapfree":
			mem.SwapFree = value
		case "swapcached":
			mem.SwapCached = value
		case "buffers":
			mem.Buffers = value
		case "active":
			mem.Active = value
		case "inactive":
			mem.Inactive = value
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	mem.Free = mem.Free + mem.Cached + mem.Buffers
	mem.Used = mem.Total - mem.Free
	mem.SwapUsed = mem.SwapTotal - mem.SwapFree

	return mem, nil
}
