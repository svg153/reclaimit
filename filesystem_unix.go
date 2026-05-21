//go:build !windows

package main

import "syscall"

func filesystemUsage(path string) (total, free, available int64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0
	}
	blockSize := int64(stat.Bsize)
	return int64(stat.Blocks) * blockSize, int64(stat.Bfree) * blockSize, int64(stat.Bavail) * blockSize
}
