package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var actionReference = regexp.MustCompile(`\buses:\s*([^\s#]+)@([^\s#]+)`)
var commitSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

func TestWorkflowActionsUseImmutableCommitSHAs(t *testing.T) {
	workflowPaths, err := filepath.Glob(filepath.Join(repoRoot(), ".github", "workflows", "*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(workflowPaths) == 0 {
		t.Fatal("no workflows found")
	}
	for _, workflowPath := range workflowPaths {
		file, err := os.Open(workflowPath)
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()
			match := actionReference.FindStringSubmatch(line)
			if match == nil || strings.HasPrefix(match[1], "./") {
				continue
			}
			if !commitSHA.MatchString(match[2]) {
				t.Errorf("%s:%d action %s is not pinned to a full commit SHA", filepath.Base(workflowPath), lineNumber, match[1])
			}
			if !strings.Contains(line, "# v") {
				t.Errorf("%s:%d pinned action has no readable version comment", filepath.Base(workflowPath), lineNumber)
			}
		}
		if err := scanner.Err(); err != nil {
			t.Error(err)
		}
		if err := file.Close(); err != nil {
			t.Error(err)
		}
	}
}

func TestDependabotUpdatesPinnedActions(t *testing.T) {
	config, err := os.ReadFile(filepath.Join(repoRoot(), ".github", "dependabot.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(config), `package-ecosystem: "github-actions"`) {
		t.Fatal("Dependabot github-actions updates are not configured")
	}
}

func TestReleasePublishesNonRootMultiPlatformContainer(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join(repoRoot(), ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflowText := string(workflow)
	for _, required := range []string{
		"docker/setup-qemu-action@",
		"docker/setup-buildx-action@",
		"platforms: linux/amd64,linux/arm64",
	} {
		if !strings.Contains(workflowText, required) {
			t.Errorf("release workflow is missing %q", required)
		}
	}

	dockerfile, err := os.ReadFile(filepath.Join(repoRoot(), "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(dockerfile), "USER nonroot:nonroot") {
		t.Fatal("runtime container must declare a non-root user")
	}
}
