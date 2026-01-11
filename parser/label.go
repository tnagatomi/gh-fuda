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
	"github.com/tnagatomi/gh-fuda/option"
	"strings"
)

func Label(input string) ([]option.Label, error) {
	inputSplit := strings.Split(input, ",")

	var labels []option.Label
	for _, label := range inputSplit {
		parts := strings.Split(label, ":")

		var name, color, description string
		switch len(parts) {
		case 1:
			// Format: name (auto-generate color)
			name = parts[0]
		case 2:
			// Format: name:color
			name = parts[0]
			color = parts[1]
		case 3:
			// Format: name:color:desc or name::desc
			name = parts[0]
			color = parts[1]
			description = parts[2]
		default:
			return nil, fmt.Errorf("invalid label format: %s", label)
		}

		// Validate name is not empty
		if name == "" {
			return nil, fmt.Errorf("label name cannot be empty")
		}

		// Validate color format if provided, otherwise generate
		if color != "" && !isHexColor(color) {
			return nil, fmt.Errorf("invalid color format: %s", color)
		}
		if color == "" {
			color = GenerateColor(name)
		}

		labels = append(labels, option.Label{
			Name:        name,
			Color:       color,
			Description: description,
		})
	}

	return labels, nil
}

func isHexColor(s string) bool {
	length := len(s)
	if length != 3 && length != 6 {
		return false
	}

	for _, char := range s {
		if !strings.ContainsRune("0123456789abcdefABCDEF", char) {
			return false
		}
	}

	return true
}
