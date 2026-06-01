package reclaimit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/svg153/reclaimit"
)

func TestRunBlackBoxControlPaths(t *testing.T) {
	originalVersion := reclaimit.Version
	reclaimit.Version = "v9.9.9-test"
	t.Cleanup(func() {
		reclaimit.Version = originalVersion
	})

	tests := []struct {
		name         string
		args         []string
		wantCode     int
		wantStdout   []string
		wantStderr   []string
		forbidStdout []string
		forbidStderr []string
	}{
		{
			name:         "help",
			args:         []string{"clean", "--help"},
			wantCode:     0,
			wantStdout:   []string{"reclaimit clean", "The command prints a deletion preview first", "--yes"},
			forbidStderr: []string{"error:"},
		},
		{
			name:         "version",
			args:         []string{"version"},
			wantCode:     0,
			wantStdout:   []string{"reclaimit v9.9.9-test"},
			forbidStderr: []string{"error:"},
		},
		{
			name:         "invalid flag value",
			args:         []string{"analyze", "--format", "json"},
			wantCode:     1,
			wantStderr:   []string{"error: unsupported format \"json\""},
			forbidStdout: []string{"Disk usage report for"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, code := runCommand(tc.args)
			if code != tc.wantCode {
				t.Fatalf("Run(%v) exit code = %d, want %d; stdout=%q stderr=%q", tc.args, code, tc.wantCode, stdout, stderr)
			}
			for _, want := range tc.wantStdout {
				if !strings.Contains(stdout, want) {
					t.Fatalf("Run(%v) stdout missing %q: %q", tc.args, want, stdout)
				}
			}
			for _, want := range tc.wantStderr {
				if !strings.Contains(stderr, want) {
					t.Fatalf("Run(%v) stderr missing %q: %q", tc.args, want, stderr)
				}
			}
			for _, forbidden := range tc.forbidStdout {
				if strings.Contains(stdout, forbidden) {
					t.Fatalf("Run(%v) stdout unexpectedly contained %q: %q", tc.args, forbidden, stdout)
				}
			}
			for _, forbidden := range tc.forbidStderr {
				if strings.Contains(stderr, forbidden) {
					t.Fatalf("Run(%v) stderr unexpectedly contained %q: %q", tc.args, forbidden, stderr)
				}
			}
		})
	}
}

func TestRunBlackBoxAnalyzeSuccess(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "demo-repo")
	target := filepath.Join(repo, "node_modules")

	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write dep.js: %v", err)
	}

	stdout, stderr, code := runCommand([]string{"analyze", "--root", root, "--min-candidate-size", "1"})
	if code != 0 {
		t.Fatalf("Run(analyze) exit code = %d; stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, want := range []string{
		"Disk usage report for " + root,
		"Top cleanup candidates",
		target,
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("Run(analyze) stdout missing %q: %q", want, stdout)
		}
	}
	if stderr != "" {
		t.Fatalf("Run(analyze) wrote to stderr: %q", stderr)
	}
}

func runCommand(args []string) (stdout string, stderr string, code int) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	code = reclaimit.Run(args, &stdoutBuf, &stderrBuf)
	return stdoutBuf.String(), stderrBuf.String(), code
}
