package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSyncCmd_MutuallyExclusiveFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "json and yaml together",
			args:        []string{"sync", "--json", "labels.json", "--yaml", "labels.yaml", "-R", "owner/repo", "--force"},
			wantErr:     true,
			errContains: "cannot be used together",
		},
		{
			name:        "json and labels together",
			args:        []string{"sync", "--json", "labels.json", "--labels", "bug:ff0000", "-R", "owner/repo", "--force"},
			wantErr:     true,
			errContains: "cannot be used together",
		},
		{
			name:        "yaml and labels together",
			args:        []string{"sync", "--yaml", "labels.yaml", "--labels", "bug:ff0000", "-R", "owner/repo", "--force"},
			wantErr:     true,
			errContains: "cannot be used together",
		},
		{
			name:        "all three together",
			args:        []string{"sync", "--json", "labels.json", "--yaml", "labels.yaml", "--labels", "bug:ff0000", "-R", "owner/repo", "--force"},
			wantErr:     true,
			errContains: "cannot be used together",
		},
		{
			name:        "no input method specified",
			args:        []string{"sync", "-R", "owner/repo", "--force"},
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
