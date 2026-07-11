package cmd

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseExcludeLabels(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        []string
		errContains string
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "comma separated labels",
			input: "good first issue, help wanted",
			want:  []string{"good first issue", "help wanted"},
		},
		{
			name:  "keeps colons in label names",
			input: "type: bug, priority: high",
			want:  []string{"type: bug", "priority: high"},
		},
		{
			name:        "rejects empty label name",
			input:       "bug,,help wanted",
			errContains: "label name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExcludeLabels(tt.input)
			if tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("parseExcludeLabels() error = %v, want containing %q", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseExcludeLabels() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseExcludeLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
