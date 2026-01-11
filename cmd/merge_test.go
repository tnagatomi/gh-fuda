package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestMergeCmd_RequiredFlags(t *testing.T) {
	// Reset flags
	fromLabel = ""
	toLabel = ""
	repos = ""
	skipConfirm = false
	dryRun = false

	var out bytes.Buffer
	rootCmd.SetArgs([]string{"merge", "-R", "owner/repo"})
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	err := rootCmd.Execute()

	if err == nil {
		t.Errorf("Execute() expected error for missing required flags, got nil")
		return
	}

	// Should fail because --from and --to are required
	if !strings.Contains(err.Error(), "required flag(s)") {
		t.Errorf("Execute() error = %v, want error containing %q", err, "required flag(s)")
	}
}

func TestMergeCmd_SameLabel(t *testing.T) {
	// Reset flags
	fromLabel = ""
	toLabel = ""
	repos = ""
	skipConfirm = false
	dryRun = false

	var out bytes.Buffer
	rootCmd.SetArgs([]string{"merge", "-R", "owner/repo", "--from", "bug", "--to", "bug"})
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	err := rootCmd.Execute()

	if err == nil {
		t.Errorf("Execute() expected error for same labels, got nil")
		return
	}

	// Should fail because from and to labels are the same
	if !strings.Contains(err.Error(), "source and target labels must be different") {
		t.Errorf("Execute() error = %v, want error containing %q", err, "source and target labels must be different")
	}
}

func TestMergeCmd_SameLabelCaseInsensitive(t *testing.T) {
	// Reset flags
	fromLabel = ""
	toLabel = ""
	repos = ""
	skipConfirm = false
	dryRun = false

	var out bytes.Buffer
	rootCmd.SetArgs([]string{"merge", "-R", "owner/repo", "--from", "Bug", "--to", "bug"})
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	err := rootCmd.Execute()

	if err == nil {
		t.Errorf("Execute() expected error for same labels (case-insensitive), got nil")
		return
	}

	// Should fail because from and to labels are the same (case-insensitive)
	if !strings.Contains(err.Error(), "source and target labels must be different") {
		t.Errorf("Execute() error = %v, want error containing %q", err, "source and target labels must be different")
	}
}
