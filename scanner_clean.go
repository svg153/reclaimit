package reclaimit

import (
	"fmt"
	"os"
)

func Clean(candidates []Candidate) (int64, error) {
	var deleted int64
	for _, candidate := range candidates {
		if err := os.RemoveAll(candidate.Path); err != nil {
			return deleted, fmt.Errorf("deleting %s: %w", candidate.Path, err)
		}
		deleted += candidate.Bytes
	}
	return deleted, nil
}

// DryRun validates that all candidate paths exist and returns the total bytes
// that would be deleted without actually removing anything.
func DryRun(candidates []Candidate) (int64, error) {
	var total int64
	for _, candidate := range candidates {
		_, err := os.Lstat(candidate.Path)
		if err != nil && os.IsNotExist(err) {
			// Path already gone — skip silently, don't fail dry-run
			continue
		}
		if err != nil {
			return total, fmt.Errorf("stat %s: %w", candidate.Path, err)
		}
		total += candidate.Bytes
	}
	return total, nil
}

// filesystemUsage is implemented per-OS in filesystem_*.go
