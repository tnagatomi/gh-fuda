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

func TestMergeCmd_Validation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     string
		errContext  string
	}{
		{
			name:       "missing required flags",
			args:       []string{"merge", "-R", "owner/repo"},
			wantErr:    "required flag(s)",
			errContext: "missing required flags",
		},
		{
			name:       "same labels",
			args:       []string{"merge", "-R", "owner/repo", "--from", "bug", "--to", "bug"},
			wantErr:    "source and target labels must be different",
			errContext: "same labels",
		},
		{
			name:       "same labels case-insensitive",
			args:       []string{"merge", "-R", "owner/repo", "--from", "Bug", "--to", "bug"},
			wantErr:    "source and target labels must be different",
			errContext: "same labels (case-insensitive)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			fromLabel = ""
			toLabel = ""
			repos = ""
			skipConfirm = false
			dryRun = false

			var out bytes.Buffer
			rootCmd.SetArgs(tt.args)
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)

			err := rootCmd.Execute()

			if err == nil {
				t.Errorf("Execute() expected error for %s, got nil", tt.errContext)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Execute() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}
