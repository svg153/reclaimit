package main

import (
	"os"

	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBinarySmoke(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "reclaimit")
	if err := runCmd("go", "build", "-o", binPath, "./cmd/reclaimit"); err != nil {
		t.Fatalf("build binary: %v", err)
	}

	versionOut, err := runCmdOutput(binPath, "version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(versionOut, "reclaimit ") {
		t.Fatalf("unexpected version output: %q", versionOut)
	}

	helpOut, err := runCmdOutput(binPath, "help", "analyze")
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	if !strings.Contains(helpOut, "reclaimit analyze") {
		t.Fatalf("unexpected help output: %q", helpOut)
	}

	root := t.TempDir()
	repo := filepath.Join(root, "repo")
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

	analyzeOut, err := runCmdOutput(binPath, "analyze", "--root", root, "--min-candidate-size", "1")
	if err != nil {
		t.Fatalf("analyze command failed: %v", err)
	}
	for _, want := range []string{"Disk usage report for", "Top cleanup candidates", target} {
		if !strings.Contains(analyzeOut, want) {
			t.Fatalf("analyze output missing %q: %q", want, analyzeOut)
		}
	}
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = "/root/reclaimit"
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCmdOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	return string(out), err
}
