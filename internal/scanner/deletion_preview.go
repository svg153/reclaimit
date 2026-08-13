package scanner

import (
	"fmt"
	"strings"
)

// RenderDeletionPreview returns a human-readable preview of deletion candidates.
func RenderDeletionPreview(candidates []Candidate) string {
	var b strings.Builder
	if len(candidates) == 0 {
		return "No cleanup candidates selected.\n"
	}

	total := sumCandidateBytes(candidates)
	fmt.Fprintf(&b, "Deleting %d candidates totaling %s:\n", len(candidates), humanizeBytes(total))
	limit := len(candidates)
	if limit > 20 {
		limit = 20
	}
	for _, candidate := range candidates[:limit] {
		fmt.Fprintf(&b, "  - %s  %s\n", humanizeBytes(candidate.Bytes), candidate.Path)
	}
	if len(candidates) > limit {
		fmt.Fprintf(&b, "  ... and %d more\n", len(candidates)-limit)
	}
	b.WriteString("\n")
	return b.String()
}

func sumCandidateBytes(candidates []Candidate) int64 {
	var total int64
	for _, c := range candidates {
		total += c.Bytes
	}
	return total
}

func humanizeBytes(bytes int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	if bytes == 0 {
		return "0 B"
	}
	idx := 0
	size := float64(bytes)
	for size >= 1024 && idx < len(units)-1 {
		size /= 1024
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%d %s", int(size), units[idx])
	}
	return fmt.Sprintf("%.1f %s", size, units[idx])
}
