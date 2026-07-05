package reclaimit

import (
	"path/filepath"
	"sort"
	"strings"
)

func pushTop(list []PathSize, item PathSize, limit int) []PathSize {
	list = append(list, item)
	sortPathSizes(list)
	if len(list) > limit {
		list = list[:limit]
	}
	return list
}

func sortPathSizes(items []PathSize) {
	sortByBytesAndPath(items, func(item PathSize) int64 {
		return item.Bytes
	}, func(item PathSize) string {
		return item.Path
	})
}

func sortCandidates(items []Candidate) {
	sortByBytesAndPath(items, func(item Candidate) int64 {
		return item.Bytes
	}, func(item Candidate) string {
		return item.Path
	})
}

func sortByBytesAndPath[T any](items []T, bytes func(T) int64, path func(T) string) {
	sort.Slice(items, func(i, j int) bool {
		if bytes(items[i]) == bytes(items[j]) {
			return path(items[i]) < path(items[j])
		}
		return bytes(items[i]) > bytes(items[j])
	})
}

// aggregateCandidates groups candidates by a key, accumulates one entry per
// key, and returns the entries sorted by the provided ordering. It is the
// shared skeleton behind summarizeCategories and summarizeGroups.
func aggregateCandidates[K comparable, V any](
	candidates []Candidate,
	keyOf func(Candidate) K,
	newEntry func(Candidate) V,
	accumulate func(*V, Candidate),
	less func(V, V) bool,
) []V {
	index := make(map[K]int)
	values := make([]V, 0)
	for _, candidate := range candidates {
		key := keyOf(candidate)
		idx, ok := index[key]
		if !ok {
			idx = len(values)
			index[key] = idx
			values = append(values, newEntry(candidate))
		}
		accumulate(&values[idx], candidate)
	}
	sort.Slice(values, func(i, j int) bool {
		return less(values[i], values[j])
	})
	return values
}

func summarizeCategories(candidates []Candidate) []CategorySummary {
	return aggregateCandidates(
		candidates,
		func(c Candidate) string { return c.CategoryKey },
		func(c Candidate) CategorySummary {
			return CategorySummary{
				Category:    c.Category,
				CategoryKey: c.CategoryKey,
				Description: c.Description,
			}
		},
		func(s *CategorySummary, c Candidate) {
			s.Bytes += c.Bytes
			s.Count++
		},
		func(a, b CategorySummary) bool {
			if a.Bytes == b.Bytes {
				return a.Category < b.Category
			}
			return a.Bytes > b.Bytes
		},
	)
}

func summarizeGroups(candidates []Candidate, limit int) []GroupSummary {
	summaries := aggregateCandidates(
		candidates,
		func(c Candidate) string { return c.Group },
		func(c Candidate) GroupSummary { return GroupSummary{Group: c.Group} },
		func(s *GroupSummary, c Candidate) {
			s.Bytes += c.Bytes
			s.Count++
			if c.ModifiedAt.After(s.ModifiedAt) {
				s.ModifiedAt = c.ModifiedAt
			}
		},
		func(a, b GroupSummary) bool {
			if a.Bytes == b.Bytes {
				return a.Group < b.Group
			}
			return a.Bytes > b.Bytes
		},
	)
	if len(summaries) > limit {
		summaries = summaries[:limit]
	}
	return summaries
}

func isGroupExcluded(candidate Candidate, excludedGroups []string) bool {
	for _, group := range excludedGroups {
		if hasPathPrefix(candidate.Group, group) || hasPathPrefix(candidate.Path, group) {
			return true
		}
	}
	return false
}

func hasPathPrefix(path, prefix string) bool {
	if path == prefix {
		return true
	}
	rel, err := filepath.Rel(prefix, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func sumBytes(items []PathSize) int64 {
	return sumBy(items, func(item PathSize) int64 {
		return item.Bytes
	})
}

func sumCandidateBytes(items []Candidate) int64 {
	return sumBy(items, func(item Candidate) int64 {
		return item.Bytes
	})
}

func sumBy[T any](items []T, bytes func(T) int64) int64 {
	var total int64
	for _, item := range items {
		total += bytes(item)
	}
	return total
}

func listToSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
