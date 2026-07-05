//go:build !windows

package filesystem

import (
	"math"
	"syscall"
)

func FilesystemUsage(path string) (total, free, available int64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0
	}
	// Field types vary across unix platforms (int32/uint32/int64/uint64), so we
	// normalise through uint64 and clamp to avoid signed-overflow on huge volumes.
	blockSize := uint64(stat.Bsize) //nolint:unconvert // type differs per platform
	blocks := uint64(stat.Blocks)   //nolint:unconvert // type differs per platform
	bfree := uint64(stat.Bfree)     //nolint:unconvert // type differs per platform
	bavail := uint64(stat.Bavail)   //nolint:unconvert // type differs per platform
	return clampUint64ToInt64(blocks * blockSize),
		clampUint64ToInt64(bfree * blockSize),
		clampUint64ToInt64(bavail * blockSize)
}

// clampUint64ToInt64 caps a uint64 at math.MaxInt64 to prevent overflow when
// converting filesystem block counts into signed byte totals.
func clampUint64ToInt64(v uint64) int64 {
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(v)
}
