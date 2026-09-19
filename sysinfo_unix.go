//go:build !windows

package main

import (
	"os/user"
	"strconv"
	"syscall"
)

func getUserAndGroup(sys any, fullPath string) (string, string) {
	stat, ok := sys.(*syscall.Stat_t)
	if !ok {
		return "", ""
	}

	uid := strconv.FormatUint(uint64(stat.Uid), 10)
	gid := strconv.FormatUint(uint64(stat.Gid), 10)

	userName := uid
	groupName := gid

	if u, err := user.LookupId(uid); err == nil {
		userName = u.Username
	}
	if g, err := user.LookupGroupId(gid); err == nil {
		groupName = g.Name
	}

	return userName, groupName
}

func isHiddenFile(filePath string) bool {
	// On Unix, hidden files are only determined by the dot prefix,
	// which is already handled in main.go
	return false
}
