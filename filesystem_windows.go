//go:build windows

package main

import "syscall"

func filesystemUsage(path string) (total, free, available int64) {
	var freeBytesAvailableToCaller, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	err := syscall.GetDiskFreeSpaceEx(syscall.StringToUTF16Ptr(path), &freeBytesAvailableToCaller, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return 0, 0, 0
	}
	return int64(totalNumberOfBytes), int64(totalNumberOfFreeBytes), int64(freeBytesAvailableToCaller)
}
