//go:build !windows

package filesystem

import (
	"testing"
)

func TestFilesystemUsage(t *testing.T) {
	total, free, available := FilesystemUsage("/")
	if total <= 0 {
		t.Errorf("expected total > 0, got %d", total)
	}
	if free < 0 || available < 0 {
		t.Errorf("free=%d, available=%d should be non-negative", free, available)
	}
	t.Logf("total=%d free=%d available=%d", total, free, available)
}

func TestFilesystemUsage_NonExistent(t *testing.T) {
	total, free, available := FilesystemUsage("/nonexistent/path/12345")
	if total != 0 || free != 0 || available != 0 {
		t.Errorf("expected 0 for nonexistent path, got total=%d free=%d available=%d", total, free, available)
	}
}

func TestClampUint64ToInt64_Normal(t *testing.T) {
	result := clampUint64ToInt64(1024)
	if result != 1024 {
		t.Errorf("expected 1024, got %d", result)
	}
}

func TestClampUint64ToInt64_Overflow(t *testing.T) {
	result := clampUint64ToInt64(18446744073709551615) // math.MaxUint64
	if result != 9223372036854775807 { // math.MaxInt64
		t.Errorf("expected math.MaxInt64, got %d", result)
	}
}

func TestClampUint64ToInt64_Zero(t *testing.T) {
	result := clampUint64ToInt64(0)
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestClampUint64ToInt64_MaxInt64(t *testing.T) {
	result := clampUint64ToInt64(9223372036854775807) // math.MaxInt64
	if result != 9223372036854775807 {
		t.Errorf("expected math.MaxInt64, got %d", result)
	}
}

func TestClampUint64ToInt64_MaxInt64PlusOne(t *testing.T) {
	result := clampUint64ToInt64(9223372036854775808) // math.MaxInt64 + 1
	if result != 9223372036854775807 {
		t.Errorf("expected math.MaxInt64, got %d", result)
	}
}
