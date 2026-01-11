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

package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tnagatomi/gh-fuda/option"
)

func TestLabelFromYAML(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		want        []option.Label
		wantErr     bool
		errContains string
	}{
		{
			name: "valid YAML with multiple labels",
			yamlContent: `- name: bug
  color: d73a4a
  description: Something isn't working
- name: enhancement
  color: a2eeef
  description: New feature or request
- name: documentation
  color: 0075ca
  description: Improvements or additions to documentation`,
			want: []option.Label{
				{Name: "bug", Color: "d73a4a", Description: "Something isn't working"},
				{Name: "enhancement", Color: "a2eeef", Description: "New feature or request"},
				{Name: "documentation", Color: "0075ca", Description: "Improvements or additions to documentation"},
			},
			wantErr: false,
		},
		{
			name: "valid YAML with label without description",
			yamlContent: `- name: bug
  color: d73a4a
- name: feature
  color: a2eeef
  description: A new feature`,
			want: []option.Label{
				{Name: "bug", Color: "d73a4a", Description: ""},
				{Name: "feature", Color: "a2eeef", Description: "A new feature"},
			},
			wantErr: false,
		},
		{
			name: "valid YAML with 3-char hex colors",
			yamlContent: `- name: bug
  color: f00
  description: Critical bug
- name: feature
  color: 0f0
  description: New feature`,
			want: []option.Label{
				{Name: "bug", Color: "f00", Description: "Critical bug"},
				{Name: "feature", Color: "0f0", Description: "New feature"},
			},
			wantErr: false,
		},
		{
			name:        "empty YAML array",
			yamlContent: `[]`,
			want:        []option.Label{},
			wantErr:     false,
		},
		{
			name: "invalid YAML syntax",
			yamlContent: `- name: bug
  color d73a4a
  description: Invalid syntax`,
			wantErr:     true,
			errContains: "failed to parse YAML",
		},
		{
			name: "label with empty name",
			yamlContent: `- name: ""
  color: d73a4a
  description: No name`,
			wantErr:     true,
			errContains: "has empty name",
		},
		{
			name: "label with missing name field",
			yamlContent: `- color: d73a4a
  description: No name field`,
			wantErr:     true,
			errContains: "has empty name",
		},
		{
			name: "label with empty color - auto generate",
			yamlContent: `- name: bug
  color: ""
  description: No color`,
			want: []option.Label{
				{Name: "bug", Color: GenerateColor("bug"), Description: "No color"},
			},
			wantErr: false,
		},
		{
			name: "label with missing color field - auto generate",
			yamlContent: `- name: bug
  description: No color field`,
			want: []option.Label{
				{Name: "bug", Color: GenerateColor("bug"), Description: "No color field"},
			},
			wantErr: false,
		},
		{
			name: "label with invalid hex color",
			yamlContent: `- name: bug
  color: gggggg
  description: Invalid hex`,
			wantErr:     true,
			errContains: "invalid color format",
		},
		{
			name: "label with invalid color length",
			yamlContent: `- name: bug
  color: ff
  description: Too short`,
			wantErr:     true,
			errContains: "invalid color format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary YAML file
			tmpDir := t.TempDir()
			yamlFile := filepath.Join(tmpDir, "labels.yaml")
			err := os.WriteFile(yamlFile, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Test the function
			got, err := LabelFromYAML(yamlFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LabelFromYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("LabelFromYAML() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("LabelFromYAML() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLabelFromYAML_FileReadError(t *testing.T) {
	// Test with non-existent file
	_, err := LabelFromYAML("/non/existent/file.yaml")
	if err == nil {
		t.Error("LabelFromYAML() should return error for non-existent file")
		return
	}
	if !strings.Contains(err.Error(), "failed to read YAML file") {
		t.Errorf("LabelFromYAML() error = %v, want error containing 'failed to read YAML file'", err)
	}
}
