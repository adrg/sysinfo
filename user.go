package sysinfo

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type User struct {
	ID       uint64
	GroupID  uint64
	Username string
	Home     string
	Shell    string
	Info     []string
}

func (u *User) Groups() ([]*Group, error) {
	groups, err := Groups()
	if err != nil {
		return nil, err
	}

	var userGroups []*Group
	for _, group := range groups {
		var g *Group
		for _, username := range group.usernames {
			if username == u.Username {
				g = group
				break
			}
		}

		if group.ID == u.ID && group.ID >= 1000 && g == nil {
			g = group
		}

		if g == nil {
			continue
		}

		userGroups = append(userGroups, g)
	}

	return userGroups, nil
}

func CurrentUser() (*User, error) {
	return UserByID(uint64(os.Getuid()))
}

func UserByID(id uint64) (*User, error) {
	users, err := Users()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, UserNotFound
}

func UserByName(username string) (*User, error) {
	users, err := Users()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, UserNotFound
}

func Users() ([]*User, error) {
	var users []*User

	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) != 7 {
			return nil, InvalidFileFormat
		}

		id, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return nil, err
		}

		groupID, err := strconv.ParseUint(fields[3], 10, 64)
		if err != nil {
			return nil, err
		}

		user := &User{
			ID:       id,
			GroupID:  groupID,
			Username: fields[0],
			Home:     fields[5],
			Shell:    fields[6],
		}

		gecosFields := strings.Split(fields[4], ",")
		for _, gecosField := range gecosFields {
			if gecosField != "" {
				user.Info = append(user.Info, gecosField)
			}
		}

		users = append(users, user)
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
