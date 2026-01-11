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

func TestCreateCmd_MutuallyExclusiveFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "json and yaml together",
			args:        []string{"create", "--json", "labels.json", "--yaml", "labels.yaml", "-R", "owner/repo"},
			wantErr:     true,
			errContains: "cannot be used together",
		},
		{
			name:        "json and labels together",
			args:        []string{"create", "--json", "labels.json", "--labels", "bug:ff0000", "-R", "owner/repo"},
			wantErr:     true,
			errContains: "cannot be used together",
		},
		{
			name:        "yaml and labels together",
			args:        []string{"create", "--yaml", "labels.yaml", "--labels", "bug:ff0000", "-R", "owner/repo"},
			wantErr:     true,
			errContains: "cannot be used together",
		},
		{
			name:        "all three together",
			args:        []string{"create", "--json", "labels.json", "--yaml", "labels.yaml", "--labels", "bug:ff0000", "-R", "owner/repo"},
			wantErr:     true,
			errContains: "cannot be used together",
		},
		{
			name:        "no input method specified",
			args:        []string{"create", "-R", "owner/repo"},
			wantErr:     true,
			errContains: "one of --labels (-l), --json, or --yaml must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			labels = ""
			jsonPath = ""
			yamlPath = ""
			repos = ""

			var out bytes.Buffer
			// Use rootCmd to get all flags properly registered
			rootCmd.SetArgs(tt.args)
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)

			err := rootCmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Execute() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}
