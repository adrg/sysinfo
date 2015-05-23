package sysinfo

import (
	"strconv"
	"strings"
)

type Load struct {
	Avg1        float64
	Avg5        float64
	Avg15       float64
	ActiveTasks uint64
	TotalTasks  uint64
	LastPID     uint64
}

func LoadAvg() (*Load, error) {
	content, err := readSingleValueFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(content)
	if len(fields) != 5 {
		return nil, ErrInvalidFileFormat
	}

	load := &Load{}
	if load.Avg1, err = strconv.ParseFloat(fields[0], 64); err != nil {
		return nil, err
	}

	if load.Avg5, err = strconv.ParseFloat(fields[1], 64); err != nil {
		return nil, err
	}

	if load.Avg15, err = strconv.ParseFloat(fields[2], 64); err != nil {
		return nil, err
	}

	tasksRatioFields := strings.Split(fields[3], "/")
	if len(tasksRatioFields) != 2 {
		return nil, ErrInvalidFileFormat
	}

	load.ActiveTasks, err = strconv.ParseUint(tasksRatioFields[0], 10, 64)
	if err != nil {
		return nil, err
	}

	load.TotalTasks, err = strconv.ParseUint(tasksRatioFields[1], 10, 64)
	if err != nil {
		return nil, err
	}

	if load.LastPID, err = strconv.ParseUint(fields[4], 10, 64); err != nil {
		return nil, err
	}

	return load, nil
}
