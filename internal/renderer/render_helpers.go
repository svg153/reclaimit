package renderer

import (
	"fmt"

	"github.com/svg153/reclaimit/internal/scanner"
	"path/filepath"
	"strings"
	"time"
)

func renderMarkdownSummary(report scanner.Report, selectionMode bool) string {
	var b strings.Builder
	b.WriteString("## Executive summary\n\n")
	b.WriteString("| Metric | Value |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| Filesystem total | %s |\n", humanizeBytes(report.FilesystemBytes))
	fmt.Fprintf(&b, "| Scanned under root | %s |\n", humanizeBytes(report.TotalBytes))
	fmt.Fprintf(&b, "| Free space | %s |\n", humanizeBytes(report.AvailableBytes))
	fmt.Fprintf(&b, "| Cleanup candidates | %d |\n", len(report.Candidates))
	fmt.Fprintf(&b, "| Potential reclaimable | %s |\n", humanizeBytes(report.CandidateBytes))
	if selectionMode {
		fmt.Fprintf(&b, "| Selected reclaimable | %s |\n", humanizeBytes(report.SelectedBytes))
	}
	return b.String()
}

func renderMarkdownCategoryPie(summaries []scanner.CategorySummary) string {
	var b strings.Builder
	b.WriteString("## Cleanup mix by category\n\n```mermaid\npie showData title Cleanup by category\n")
	for _, summary := range summaries {
		fmt.Fprintf(&b, "\"%s\" : %.1f\n", summary.CategoryKey, bytesToMiB(summary.Bytes))
	}
	b.WriteString("```\n")
	return b.String()
}

func renderMarkdownTopGroupsChart(groups []scanner.GroupSummary) string {
	var b strings.Builder
	limited := groups
	if len(limited) > 8 {
		limited = limited[:8]
	}
	b.WriteString("## Top reclaimable groups\n\n```mermaid\nxychart-beta\n")
	b.WriteString("    title \"Top reclaimable groups (GiB)\"\n")
	b.WriteString("    x-axis [")
	for i, group := range limited {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "\"%s\"", filepath.Base(group.Group))
	}
	b.WriteString("]\n    y-axis \"GiB\" 0 --> ")
	maxGiB := 1.0
	for _, group := range limited {
		if gib := bytesToGiB(group.Bytes); gib > maxGiB {
			maxGiB = gib
		}
	}
	fmt.Fprintf(&b, "%.1f\n", maxGiB+0.5)
	b.WriteString("    bar [")
	for i, group := range limited {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%.2f", bytesToGiB(group.Bytes))
	}
	b.WriteString("]\n```\n")
	return b.String()
}

func renderPlantUMLOverview(groups []scanner.GroupSummary) string {
	var b strings.Builder
	limited := groups
	if len(limited) > 8 {
		limited = limited[:8]
	}
	b.WriteString("## PlantUML group map\n\n```plantuml\n@startmindmap\n* reclaimit\n")
	for _, group := range limited {
		fmt.Fprintf(&b, "** %s (%s)\n", escapePlant(filepath.Base(group.Group)), humanizeBytes(group.Bytes))
	}
	b.WriteString("@endmindmap\n```\n")
	return b.String()
}

func renderMarkdownDetails(title string, header string, rows []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<details>\n<summary>%s</summary>\n\n", title)
	b.WriteString(header)
	for _, row := range rows {
		b.WriteString(row)
	}
	b.WriteString("\n</details>\n")
	return b.String()
}

func pathSizeRows(items []scanner.PathSize) []string {
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, fmt.Sprintf("| %s | `%s` |\n", humanizeBytes(item.Bytes), escapeMarkdownCell(item.Path)))
	}
	return rows
}

func categoryRows(items []scanner.CategorySummary) []string {
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, fmt.Sprintf("| %s | `%s` | %d | %s |\n", humanizeBytes(item.Bytes), item.CategoryKey, item.Count, escapeMarkdownCell(item.Description)))
	}
	return rows
}

func groupRows(items []scanner.GroupSummary) []string {
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, fmt.Sprintf("| %s | %d | `%s` | %s |\n", humanizeBytes(item.Bytes), item.Count, escapeMarkdownCell(item.Group), humanizeTimestamp(item.ModifiedAt)))
	}
	return rows
}

func candidateRows(items []scanner.Candidate, max int) []string {
	limited := limitCandidates(items, max)
	rows := make([]string, 0, len(limited))
	for _, item := range limited {
		rows = append(rows, fmt.Sprintf("| %s | `%s` | %s | %s | `%s` | `%s` |\n", humanizeBytes(item.Bytes), item.CategoryKey, candidateKind(item), humanizeTimestamp(item.ModifiedAt), escapeMarkdownCell(item.Group), escapeMarkdownCell(item.Path)))
	}
	return rows
}

func escapeMarkdownCell(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

func humanizeTimestamp(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.UTC().Format("2006-01-02 15:04")
}

func candidateKind(item scanner.Candidate) string {
	if item.IsDir {
		return "dir"
	}
	return "file"
}

func bytesToMiB(size int64) float64 { return float64(size) / 1024.0 / 1024.0 }
func bytesToGiB(size int64) float64 { return bytesToMiB(size) / 1024.0 }
