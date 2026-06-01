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

// filesystemUsage is implemented per-OS in filesystem_*.go
