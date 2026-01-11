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

	"github.com/tnagatomi/gh-fuda/option"
)

func TestLabelFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonContent string
		want        []option.Label
		wantErr     bool
		errContains string
	}{
		{
			name: "valid JSON with multiple labels",
			jsonContent: `[
				{
					"name": "Status: Completed",
					"color": "2E7D32",
					"description": "Work was started and finished"
				},
				{
					"name": "Status: In Progress",
					"color": "666699",
					"description": "Work was started, is actively being worked on"
				},
				{
					"name": "Priority: High",
					"color": "FF8C00",
					"description": ""
				}
			]`,
			want: []option.Label{
				{Name: "Status: Completed", Color: "2E7D32", Description: "Work was started and finished"},
				{Name: "Status: In Progress", Color: "666699", Description: "Work was started, is actively being worked on"},
				{Name: "Priority: High", Color: "FF8C00", Description: ""},
			},
			wantErr: false,
		},
		{
			name: "valid JSON with single label",
			jsonContent: `[
				{
					"name": "bug",
					"color": "d73a4a",
					"description": "Something isn't working"
				}
			]`,
			want: []option.Label{
				{Name: "bug", Color: "d73a4a", Description: "Something isn't working"},
			},
			wantErr: false,
		},
		{
			name:        "empty JSON array",
			jsonContent: `[]`,
			want:        []option.Label{},
			wantErr:     false,
		},
		{
			name: "label with empty name",
			jsonContent: `[
				{
					"name": "",
					"color": "FF0000",
					"description": "Test"
				}
			]`,
			want:        nil,
			wantErr:     true,
			errContains: "empty name",
		},
		{
			name: "label with empty color - auto generate",
			jsonContent: `[
				{
					"name": "test",
					"color": "",
					"description": "Test"
				}
			]`,
			want: []option.Label{
				{Name: "test", Color: GenerateColor("test"), Description: "Test"},
			},
			wantErr: false,
		},
		{
			name: "label without color field - auto generate",
			jsonContent: `[
				{
					"name": "bug",
					"description": "Bug report"
				}
			]`,
			want: []option.Label{
				{Name: "bug", Color: GenerateColor("bug"), Description: "Bug report"},
			},
			wantErr: false,
		},
		{
			name: "label with invalid color format",
			jsonContent: `[
				{
					"name": "test",
					"color": "GGGGGG",
					"description": "Test"
				}
			]`,
			want:        nil,
			wantErr:     true,
			errContains: "invalid color format",
		},
		{
			name: "label with invalid color length",
			jsonContent: `[
				{
					"name": "test",
					"color": "FF00",
					"description": "Test"
				}
			]`,
			want:        nil,
			wantErr:     true,
			errContains: "invalid color format",
		},
		{
			name:        "invalid JSON format",
			jsonContent: `{"invalid": "json"}`,
			want:        nil,
			wantErr:     true,
			errContains: "failed to parse JSON",
		},
		{
			name:        "malformed JSON",
			jsonContent: `[{invalid json}]`,
			want:        nil,
			wantErr:     true,
			errContains: "failed to parse JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			jsonFile := filepath.Join(tmpDir, "labels.json")
			err := os.WriteFile(jsonFile, []byte(tt.jsonContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			got, err := LabelFromJSON(jsonFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LabelFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("LabelFromJSON() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("LabelFromJSON() returned %d labels, want %d", len(got), len(tt.want))
				return
			}

			for i, label := range got {
				if label.Name != tt.want[i].Name ||
					label.Color != tt.want[i].Color ||
					label.Description != tt.want[i].Description {
					t.Errorf("LabelFromJSON() label[%d] = %+v, want %+v", i, label, tt.want[i])
				}
			}
		})
	}
}

func TestLabelFromJSON_FileNotFound(t *testing.T) {
	_, err := LabelFromJSON("/non/existent/file.json")
	if err == nil {
		t.Error("LabelFromJSON() expected error for non-existent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read JSON file") {
		t.Errorf("LabelFromJSON() error = %v, want error containing 'failed to read JSON file'", err)
	}
}
