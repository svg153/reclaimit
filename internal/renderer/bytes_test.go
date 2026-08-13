package renderer

import (
	"testing"
)

func TestBytesToMiB(t *testing.T) {
	tests := []struct {
		input    int64
		expected float64
	}{
		{0, 0},
		{1024 * 1024, 1},
		{1024, 1.0 / 1024.0},
		{1024 * 1024 * 100, 100},
	}
	for _, tt := range tests {
		result := bytesToMiB(tt.input)
		if result != tt.expected {
			t.Errorf("bytesToMiB(%d) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestBytesToGiB(t *testing.T) {
	tests := []struct {
		input    int64
		expected float64
	}{
		{0, 0},
		{1024 * 1024 * 1024, 1},
		{1024 * 1024 * 1024 * 2, 2},
	}
	for _, tt := range tests {
		result := bytesToGiB(tt.input)
		if result != tt.expected {
			t.Errorf("bytesToGiB(%d) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}
