//go:build windows

package main

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func getUserAndGroup(sys any, fullPath string) (string, string) {
	sd, err := windows.GetNamedSecurityInfo(
		fullPath,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION,
	)
	if err != nil {
		return "", ""
	}

	owner, _, err := sd.Owner()
	if err != nil {
		return "", ""
	}

	account, _, _, err := owner.LookupAccount("")
	if err != nil {
		return "", ""
	}

	return account, ""
}

func isHiddenFile(filePath string) bool {
	pointer, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return false
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		// If we can't access the file, we can't check its attributes safely
		// We avoid panicking here.
		return false
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}
