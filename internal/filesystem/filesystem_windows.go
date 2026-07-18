//go:build windows

package filesystem

import (
	"math"
	"syscall"
	"unsafe"
)

var (
	kernel32DLL             = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceExWProc = kernel32DLL.NewProc("GetDiskFreeSpaceExW")
)

// filesystemUsage reads total, free and available bytes using the native
// GetDiskFreeSpaceExW API. On failures it returns zeros to preserve the CLI
// best-effort behavior used on other platforms.
func FilesystemUsage(path string) (total, free, available int64) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, 0
	}

	var freeAvailable uint64
	var totalBytes uint64
	var totalFree uint64

	r1, _, _ := getDiskFreeSpaceExWProc.Call(
		uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&freeAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if r1 == 0 {
		return 0, 0, 0
	}

	return clampUint64ToInt64(totalBytes), clampUint64ToInt64(totalFree), clampUint64ToInt64(freeAvailable)
}

func clampUint64ToInt64(v uint64) int64 {
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(v)
}
