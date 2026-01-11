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
	"fmt"
	"os"

	"github.com/tnagatomi/gh-fuda/option"
	"gopkg.in/yaml.v3"
)

// YAMLLabel represents the YAML structure for a label
type YAMLLabel struct {
	Name        string `yaml:"name"`
	Color       string `yaml:"color"`
	Description string `yaml:"description"`
}

// LabelFromYAML parses labels from a YAML file
func LabelFromYAML(path string) ([]option.Label, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}

	var yamlLabels []YAMLLabel
	if err := yaml.Unmarshal(data, &yamlLabels); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	labels := make([]option.Label, 0, len(yamlLabels))
	for i, yl := range yamlLabels {
		if yl.Name == "" {
			return nil, fmt.Errorf("label at index %d has empty name", i)
		}

		// Validate color format if provided, otherwise generate
		color := yl.Color
		if color != "" && !isHexColor(color) {
			return nil, fmt.Errorf("label %q has invalid color format: %s", yl.Name, color)
		}
		if color == "" {
			color = GenerateColor(yl.Name)
		}

		labels = append(labels, option.Label{
			Name:        yl.Name,
			Color:       color,
			Description: yl.Description,
		})
	}

	return labels, nil
}
