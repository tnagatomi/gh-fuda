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
	"fmt"
	"io"
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
