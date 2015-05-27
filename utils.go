package sysinfo

import (
	"errors"
	"io/ioutil"
	"strings"
)

// #include <unistd.h>
import "C"

var (
	TicksPerSecond uint64

	ErrUserNotFound      = errors.New("user not found")
	ErrGroupNotFound     = errors.New("group not found")
	ErrInvalidFileFormat = errors.New("invalid file format")
)

func init() {
	TicksPerSecond = uint64(C.sysconf(C._SC_CLK_TCK))
}

func readSingleValueFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}
