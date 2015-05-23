package sysinfo

import (
	"os/exec"
	"strconv"
	"strings"
)

type Node struct {
	Hostname      string
	Domain        string
	Architecture  string
	KernelName    string
	KernelRelease string
	KernelVersion string
	Uptime        float64
}

func NodeInfo() (*Node, error) {
	node := &Node{}

	var err error
	if node.Hostname, err = Hostname(); err != nil {
		return nil, err
	}

	if node.Domain, err = Domain(); err != nil {
		return nil, err
	}

	if node.Architecture, err = Architecture(); err != nil {
		return nil, err
	}

	if node.KernelName, err = KernelName(); err != nil {
		return nil, err
	}

	if node.KernelRelease, err = KernelRelease(); err != nil {
		return nil, err
	}

	if node.KernelVersion, err = KernelVersion(); err != nil {
		return nil, err
	}

	if node.Uptime, err = Uptime(); err != nil {
		return nil, err
	}

	return node, nil
}

func Hostname() (string, error) {
	return readSingleValueFile("/proc/sys/kernel/hostname")
}

func Domain() (string, error) {
	domain, err := readSingleValueFile("/proc/sys/kernel/domainname")
	if domain == "(none)" {
		domain = ""
	}

	return domain, err
}

func KernelName() (string, error) {
	return readSingleValueFile("/proc/sys/kernel/ostype")
}

func KernelRelease() (string, error) {
	return readSingleValueFile("/proc/sys/kernel/osrelease")
}

func KernelVersion() (string, error) {
	return readSingleValueFile("/proc/sys/kernel/version")
}

func Uptime() (float64, error) {
	content, err := readSingleValueFile("/proc/uptime")
	if err != nil {
		return 0.0, err
	}

	fields := strings.Fields(content)
	if len(fields) != 2 {
		return 0.0, ErrInvalidFileFormat
	}

	return strconv.ParseFloat(fields[0], 64)
}

func Architecture() (string, error) {
	output, err := exec.Command("uname", "-m").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
