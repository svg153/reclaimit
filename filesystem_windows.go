//go:build windows

package main

// filesystemUsage intentionally returns zeros in the cross-compile
// environment where syscall helpers may not be available. This keeps
// release matrix builds portable. At runtime on native Windows the
// values can be implemented to call the Win32 API if desired.
func filesystemUsage(path string) (total, free, available int64) {
	return 0, 0, 0
}
