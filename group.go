package sysinfo

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Group struct {
	ID   uint64
	Name string

	usernames []string
}

func (g *Group) Users() ([]*User, error) {
	users, err := Users()
	if err != nil {
		return nil, err
	}

	var groupUsers []*User
	for _, user := range users {
		var u *User
		for _, username := range g.usernames {
			if username != user.Username {
				continue
			}

			u = user
			break
		}

		if user.ID == g.ID && user.ID >= 1000 && u == nil {
			u = user
		}

		if u == nil {
			continue
		}

		groupUsers = append(groupUsers, u)
	}

	return groupUsers, nil
}

func GroupByID(id uint64) (*Group, error) {
	groups, err := Groups()
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if group.ID == id {
			return group, nil
		}
	}

	return nil, GroupNotFound
}

func GroupByName(name string) (*Group, error) {
	groups, err := Groups()
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if group.Name == name {
			return group, nil
		}
	}

	return nil, GroupNotFound
}

func Groups() ([]*Group, error) {
	var groups []*Group

	file, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) != 4 {
			return nil, InvalidFileFormat
		}

		id, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return nil, err
		}

		group := &Group{
			ID:   id,
			Name: fields[0],
		}

		names := strings.Split(fields[3], ",")
		for _, name := range names {
			if name != "" {
				group.usernames = append(group.usernames, name)
			}
		}

		groups = append(groups, group)
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}
