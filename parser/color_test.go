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
	"testing"
)

func TestGenerateColor(t *testing.T) {
	tests := []struct {
		name      string
		labelName string
		want      string
	}{
		{
			name:      "empty name returns default gray",
			labelName: "",
			want:      "cccccc",
		},
		{
			name:      "bug label",
			labelName: "bug",
			want:      "ecb53f",
		},
		{
			name:      "enhancement label",
			labelName: "enhancement",
			want:      "ca9626",
		},
		{
			name:      "unicode label",
			labelName: "バグ",
			want:      "468944",
		},
		{
			name:      "label with spaces",
			labelName: "good first issue",
			want:      "07c533",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateColor(tt.labelName)

			// Verify length is always 6
			if len(got) != 6 {
				t.Errorf("GenerateColor() returned %d chars, want 6", len(got))
			}

			// Verify valid hex format
			if !isHexColor(got) {
				t.Errorf("GenerateColor() = %q, not valid hex", got)
			}

			// Verify determinism
			got2 := GenerateColor(tt.labelName)
			if got != got2 {
				t.Errorf("GenerateColor() not deterministic: first=%q, second=%q", got, got2)
			}

			// Verify expected value
			if got != tt.want {
				t.Errorf("GenerateColor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateColor_DifferentNamesProduceDifferentColors(t *testing.T) {
	names := []string{"bug", "enhancement", "feature", "documentation", "help wanted"}
	colors := make(map[string]string)

	for _, name := range names {
		color := GenerateColor(name)
		if existing, found := colors[color]; found {
			t.Errorf("GenerateColor() collision: %q and %q both produce %q", name, existing, color)
		}
		colors[color] = name
	}
}
