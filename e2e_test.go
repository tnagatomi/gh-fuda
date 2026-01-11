//go:build e2e

/*
Copyright © 2025 Takayuki Nagatomi <tnagatomi@okweird.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// E2E tests that execute the actual CLI binary.
// These tests require:
// - GH_TOKEN or GITHUB_TOKEN environment variable
// - Test repositories: tnagatomi/gh-fuda-test-1, tnagatomi/gh-fuda-test-2
//
// Run with: go test -tags=e2e -v
//
// IMPORTANT: These tests should run with concurrency: 1 in CI to avoid conflicts.

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

const (
	testRepo1 = "tnagatomi/gh-fuda-test-1"
	testRepo2 = "tnagatomi/gh-fuda-test-2"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary before running tests
	cmd := exec.Command("go", "build", "-o", "gh-fuda-test-binary", ".")
	if err := cmd.Run(); err != nil {
		panic("Failed to build binary: " + err.Error())
	}
	binaryPath = "./gh-fuda-test-binary"

	code := m.Run()

	// Cleanup binary
	os.Remove(binaryPath)

	os.Exit(code)
}

func skipIfNoToken(t *testing.T) {
	t.Helper()
	if os.Getenv("GH_TOKEN") == "" && os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("Skipping E2E test: GH_TOKEN or GITHUB_TOKEN not set")
	}
}

func TestE2E_List_EmptyRepository(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)

	stdout := mustRunCmd(t, "list", "-R", testRepo1)

	if !strings.Contains(stdout, "has no labels") {
		t.Errorf("Expected 'has no labels' message, got: %s", stdout)
	}
}

func TestE2E_Create(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	// Create labels
	stdout := mustRunCmd(t, "create", "-R", testRepo1, "-l", "bug:ff0000:Bug label,enhancement:00ff00:Enhancement")

	if !strings.Contains(stdout, `Created label "bug"`) {
		t.Errorf("Expected 'Created label \"bug\"', got: %s", stdout)
	}
	if !strings.Contains(stdout, `Created label "enhancement"`) {
		t.Errorf("Expected 'Created label \"enhancement\"', got: %s", stdout)
	}

	// Verify with list
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if !strings.Contains(stdout, "bug") {
		t.Errorf("Expected 'bug' label in list, got: %s", stdout)
	}
	if !strings.Contains(stdout, "enhancement") {
		t.Errorf("Expected 'enhancement' label in list, got: %s", stdout)
	}
}

func TestE2E_Create_DryRun(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)

	stdout := mustRunCmd(t, "create", "-R", testRepo1, "-l", "test-label:aabbcc", "--dry-run")

	if !strings.Contains(stdout, "Would create label") {
		t.Errorf("Expected 'Would create label', got: %s", stdout)
	}

	// Verify label was not actually created
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if strings.Contains(stdout, "test-label") {
		t.Errorf("Label should not have been created in dry-run mode")
	}
}

func TestE2E_Delete(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	// Create label first
	mustRunCmd(t, "create", "-R", testRepo1, "-l", "to-delete:ff0000")

	// Delete label
	stdout := mustRunCmd(t, "delete", "-R", testRepo1, "-l", "to-delete", "--force")

	if !strings.Contains(stdout, `Deleted label "to-delete"`) {
		t.Errorf("Expected 'Deleted label \"to-delete\"', got: %s", stdout)
	}

	// Verify deletion
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if strings.Contains(stdout, "to-delete") {
		t.Errorf("Label should have been deleted")
	}
}

func TestE2E_Delete_DryRun(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	// Create label first
	mustRunCmd(t, "create", "-R", testRepo1, "-l", "keep-me:ff0000")

	// Dry-run delete
	stdout := mustRunCmd(t, "delete", "-R", testRepo1, "-l", "keep-me", "--dry-run")

	if !strings.Contains(stdout, "Would delete label") {
		t.Errorf("Expected 'Would delete label', got: %s", stdout)
	}

	// Verify label still exists
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if !strings.Contains(stdout, "keep-me") {
		t.Errorf("Label should still exist after dry-run delete")
	}
}

func TestE2E_Sync(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	// Create initial labels
	mustRunCmd(t, "create", "-R", testRepo1, "-l", "old-label:aaaaaa,keep-label:bbbbbb")

	// Sync with new set of labels
	stdout := mustRunCmd(t, "sync", "-R", testRepo1, "-l", "keep-label:cccccc:Updated,new-label:dddddd", "--force")

	// Should delete old-label, update keep-label, create new-label
	if !strings.Contains(stdout, `Deleted label "old-label"`) {
		t.Errorf("Expected old-label to be deleted, got: %s", stdout)
	}
	if !strings.Contains(stdout, `Updated label "keep-label"`) {
		t.Errorf("Expected keep-label to be updated, got: %s", stdout)
	}
	if !strings.Contains(stdout, `Created label "new-label"`) {
		t.Errorf("Expected new-label to be created, got: %s", stdout)
	}
}

func TestE2E_Sync_DryRun(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	// Create initial label
	mustRunCmd(t, "create", "-R", testRepo1, "-l", "existing:aaaaaa")

	// Dry-run sync
	stdout := mustRunCmd(t, "sync", "-R", testRepo1, "-l", "new-only:bbbbbb", "--dry-run")

	if !strings.Contains(stdout, "Would delete") {
		t.Errorf("Expected 'Would delete', got: %s", stdout)
	}
	if !strings.Contains(stdout, "Would create") {
		t.Errorf("Expected 'Would create', got: %s", stdout)
	}

	// Verify nothing changed
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if !strings.Contains(stdout, "existing") {
		t.Errorf("Original label should still exist")
	}
	if strings.Contains(stdout, "new-only") {
		t.Errorf("New label should not have been created")
	}
}

func TestE2E_Empty(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)

	// Create some labels
	mustRunCmd(t, "create", "-R", testRepo1, "-l", "label1:aa0000,label2:00aa00,label3:0000aa")

	// Empty repository
	stdout := mustRunCmd(t, "empty", "-R", testRepo1, "--force")

	if !strings.Contains(stdout, `Deleted label "label1"`) {
		t.Errorf("Expected label1 to be deleted, got: %s", stdout)
	}

	// Verify all labels deleted
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if !strings.Contains(stdout, "has no labels") {
		t.Errorf("Expected empty repository, got: %s", stdout)
	}
}

func TestE2E_Empty_DryRun(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	// Create some labels
	mustRunCmd(t, "create", "-R", testRepo1, "-l", "keep1:aa0000,keep2:00aa00")

	// Dry-run empty
	stdout := mustRunCmd(t, "empty", "-R", testRepo1, "--dry-run")

	if !strings.Contains(stdout, "Would delete") {
		t.Errorf("Expected 'Would delete', got: %s", stdout)
	}

	// Verify labels still exist
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if !strings.Contains(stdout, "keep1") || !strings.Contains(stdout, "keep2") {
		t.Errorf("Labels should still exist after dry-run empty")
	}
}

func TestE2E_Merge(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	// Create source and target labels
	mustRunCmd(t, "create", "-R", testRepo1, "-l", "old-name:ff0000,new-name:00ff00")

	// Merge (no issues with labels, so just label deletion)
	stdout := mustRunCmd(t, "merge", "-R", testRepo1, "--from", "old-name", "--to", "new-name", "-y")

	if !strings.Contains(stdout, `Deleted label "old-name"`) {
		t.Errorf("Expected source label to be deleted, got: %s", stdout)
	}

	// Verify source label deleted, target remains
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if strings.Contains(stdout, "old-name") {
		t.Errorf("Source label should have been deleted")
	}
	if !strings.Contains(stdout, "new-name") {
		t.Errorf("Target label should still exist")
	}
}

func TestE2E_Merge_DryRun(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	// Create source and target labels
	mustRunCmd(t, "create", "-R", testRepo1, "-l", "source:ff0000,target:00ff00")

	// Dry-run merge
	stdout := mustRunCmd(t, "merge", "-R", testRepo1, "--from", "source", "--to", "target", "-y", "--dry-run")

	if !strings.Contains(stdout, "Would delete") {
		t.Errorf("Expected 'Would delete', got: %s", stdout)
	}

	// Verify both labels still exist
	stdout = mustRunCmd(t, "list", "-R", testRepo1)

	if !strings.Contains(stdout, "source") {
		t.Errorf("Source label should still exist after dry-run")
	}
	if !strings.Contains(stdout, "target") {
		t.Errorf("Target label should still exist after dry-run")
	}
}

func TestE2E_MultipleRepositories(t *testing.T) {
	skipIfNoToken(t)
	cleanupRepos(t)
	t.Cleanup(func() { cleanupRepos(t) })

	repos := testRepo1 + "," + testRepo2

	// Create labels in both repos
	stdout := mustRunCmd(t, "create", "-R", repos, "-l", "shared:abcdef")

	if !strings.Contains(stdout, testRepo1) || !strings.Contains(stdout, testRepo2) {
		t.Errorf("Expected both repos in output, got: %s", stdout)
	}

	// Verify both repos have the label
	stdout1 := mustRunCmd(t, "list", "-R", testRepo1)
	stdout2 := mustRunCmd(t, "list", "-R", testRepo2)

	if !strings.Contains(stdout1, "shared") {
		t.Errorf("repo1 should have 'shared' label")
	}
	if !strings.Contains(stdout2, "shared") {
		t.Errorf("repo2 should have 'shared' label")
	}
}

func runCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func mustRunCmd(t *testing.T, args ...string) string {
	t.Helper()
	stdout, stderr, err := runCmd(t, args...)
	if err != nil {
		t.Fatalf("Command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	return stdout
}

func cleanupRepos(t *testing.T) {
	t.Helper()
	// Empty both test repositories
	runCmd(t, "empty", "-R", testRepo1, "--force")
	runCmd(t, "empty", "-R", testRepo2, "--force")
}