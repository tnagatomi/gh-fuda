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
	"encoding/json"
	"fmt"
	"os"

	"github.com/tnagatomi/gh-fuda/option"
)

// JSONLabel represents the JSON structure for a label
type JSONLabel struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// LabelFromJSON parses labels from a JSON file
func LabelFromJSON(path string) ([]option.Label, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %v", err)
	}

	var jsonLabels []JSONLabel
	if err := json.Unmarshal(data, &jsonLabels); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	var labels []option.Label
	for i, jl := range jsonLabels {
		if jl.Name == "" {
			return nil, fmt.Errorf("label at index %d has empty name", i)
		}

		if jl.Color == "" {
			return nil, fmt.Errorf("label %q has empty color", jl.Name)
		}

		if !isHexColor(jl.Color) {
			return nil, fmt.Errorf("label %q has invalid color format: %s", jl.Name, jl.Color)
		}

		labels = append(labels, option.Label{
			Name:        jl.Name,
			Color:       jl.Color,
			Description: jl.Description,
		})
	}

	return labels, nil
}
