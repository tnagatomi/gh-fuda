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
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/tnagatomi/gh-fuda/option"
	"github.com/tnagatomi/gh-fuda/parser"
)

// confirm asks user to really execute a command
func confirm(in io.Reader, out io.Writer) (bool, error) {
	_, _ = fmt.Fprintf(out, "Are you sure you want to do this? (y/n): ")

	reader := bufio.NewReader(in)
	s, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read: %v", err)
	}
	if s == "y\n" {
		return true, nil
	}
	return false, nil
}

// parseLabelsInput validates and parses label input from various sources (--labels, --json, or --yaml flags)
func parseLabelsInput(labels, jsonPath, yamlPath string) ([]option.Label, error) {
	// Check that only one input method is specified
	inputCount := 0
	if labels != "" {
		inputCount++
	}
	if jsonPath != "" {
		inputCount++
	}
	if yamlPath != "" {
		inputCount++
	}
	
	if inputCount > 1 {
		return nil, errors.New("--labels (-l), --json, and --yaml cannot be used together")
	}
	if inputCount == 0 {
		return nil, errors.New("one of --labels (-l), --json, or --yaml must be specified")
	}

	var labelList []option.Label
	var err error
	if jsonPath != "" {
		labelList, err = parser.LabelFromJSON(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON file: %v", err)
		}
	} else if yamlPath != "" {
		labelList, err = parser.LabelFromYAML(yamlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML file: %v", err)
		}
	} else {
		labelList, err = parser.Label(labels)
		if err != nil {
			return nil, fmt.Errorf("failed to parse labels option: %v", err)
		}
	}

	return labelList, nil
}
