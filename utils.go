package sysinfo

import (
	"errors"
	"io/ioutil"
	"strings"
)

var (
	UserNotFound      error = errors.New("User not found")
	GroupNotFound     error = errors.New("Group not found")
	InvalidFileFormat error = errors.New("Invalid file format")
)

func readSingleValueFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}
