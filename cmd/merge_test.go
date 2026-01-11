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
