package cli

import (
	"fmt"

	"github.com/svg153/reclaimit/internal/scanner"
	"io"
	"strings"
)

func exitf(w io.Writer, format string, args ...any) int {
	_, _ = fmt.Fprintf(w, format, args...)
	return 1
}

func writeString(w io.Writer, value string) error {
	_, err := io.WriteString(w, value)
	return err
}

func writef(w io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(w, format, args...)
	return err
}

func RenderDeletionPreview(candidates []scanner.Candidate) string {
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

func sumCandidateBytes(candidates []scanner.Candidate) int64 {
	var total int64
	for _, c := range candidates {
		total += c.Bytes
	}
	return total
}
