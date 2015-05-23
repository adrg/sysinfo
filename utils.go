package sysinfo

import (
	"errors"
	"io/ioutil"
	"strings"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrGroupNotFound     = errors.New("group not found")
	ErrInvalidFileFormat = errors.New("invalid file format")
)

func readSingleValueFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}
