package doctor

import (
	"errors"
	"strings"
	"testing"
)

func TestRunDoctorReturnsUsableReport(t *testing.T) {
	report := Run(Options{
		Version:     "1.2.3",
		Args0:       "/usr/local/bin/reclaimit",
		Getwd:       func() (string, error) { return "/private/project", nil },
		UserHomeDir: func() (string, error) { return "/home/alice", nil },
		LookPath:    func(string) (string, error) { return "/usr/local/bin/reclaimit", nil },
	})

	if got := ExitCode(report); got != 0 {
		t.Fatalf("ExitCode = %d, want 0", got)
	}
	rendered := Render(report)
	for _, want := range []string{
		"reclaimit doctor",
		"version: 1.2.3",
		"[ok] working-directory: current directory is readable",
		"[ok] home-directory: home directory detected",
		"[ok] path: reclaimit is reachable from PATH",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("Render() missing %q:\n%s", want, rendered)
		}
	}
	for _, forbidden := range []string{"/private/project", "/home/alice", "/usr/local/bin/reclaimit"} {
		if strings.Contains(rendered, forbidden) {
			t.Fatalf("Render() leaked private path %q:\n%s", forbidden, rendered)
		}
	}
}

func TestRunDoctorFailsWhenWorkingDirectoryIsUnreadable(t *testing.T) {
	report := Run(Options{
		Version:     "dev",
		Args0:       "reclaimit",
		Getwd:       func() (string, error) { return "", errors.New("permission denied") },
		UserHomeDir: func() (string, error) { return "/home/alice", nil },
		LookPath:    func(string) (string, error) { return "/usr/bin/reclaimit", nil },
	})

	if got := ExitCode(report); got != 1 {
		t.Fatalf("ExitCode = %d, want 1", got)
	}
	rendered := Render(report)
	if !strings.Contains(rendered, "[fail] working-directory: current directory is not readable: permission denied") {
		t.Fatalf("Render() missing working-directory failure:\n%s", rendered)
	}
}

func TestRunDoctorWarnsWhenBinaryIsMissingFromPath(t *testing.T) {
	report := Run(Options{
		Version:     "",
		Args0:       "",
		Getwd:       func() (string, error) { return "/workspace", nil },
		UserHomeDir: func() (string, error) { return "", errors.New("home unavailable") },
		LookPath:    func(string) (string, error) { return "", errors.New("not found") },
	})

	if got := ExitCode(report); got != 0 {
		t.Fatalf("ExitCode = %d, want 0", got)
	}
	rendered := Render(report)
	for _, want := range []string{
		"version: unknown",
		"[warn] home-directory: home directory could not be detected: home unavailable",
		"[warn] path: reclaimit is not reachable from PATH",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("Render() missing %q:\n%s", want, rendered)
		}
	}
}

func TestBinaryNameFallsBackForBlankInputs(t *testing.T) {
	for _, input := range []string{"", "   ", "/", `\`} {
		if got := binaryName(input); got != "reclaimit" {
			t.Fatalf("binaryName(%q) = %q, want reclaimit", input, got)
		}
	}
}

func TestBinaryNameStripsUnixAndWindowsDirectories(t *testing.T) {
	tests := map[string]string{
		"/opt/bin/reclaimit":      "reclaimit",
		`C:\Tools\reclaimit.exe`:  "reclaimit.exe",
		"relative/path/reclaimit": "reclaimit",
		`relative\path\reclaimit`: "reclaimit",
		"already-on-path":         "already-on-path",
	}

	for input, want := range tests {
		if got := binaryName(input); got != want {
			t.Fatalf("binaryName(%q) = %q, want %q", input, got, want)
		}
	}
}
