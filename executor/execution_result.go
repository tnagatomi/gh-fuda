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
package executor

import (
	"fmt"

	"github.com/tnagatomi/gh-fuda/api"
)

// RepoResult represents the result of operations on a repository
type RepoResult struct {
	Repo   string
	Errors []error
}

// ExecutionResult collects and reports the results of execution
type ExecutionResult struct {
	results map[string]*RepoResult
}

// NewExecutionResult creates a new ExecutionResult
func NewExecutionResult() *ExecutionResult {
	return &ExecutionResult{
		results: make(map[string]*RepoResult),
	}
}

// AddRepoResult adds a repository result to the execution result
func (er *ExecutionResult) AddRepoResult(result *RepoResult) {
	er.results[result.Repo] = result
}

// RepoResult returns the result for a specific repository, or nil if not found
func (er *ExecutionResult) RepoResult(repoName string) *RepoResult {
	return er.results[repoName]
}

// HasErrors returns true if there are any errors
func (er *ExecutionResult) HasErrors() bool {
	for _, result := range er.results {
		if len(result.Errors) > 0 {
			return true
		}
	}
	return false
}

// IsNotFoundError checks if the error is a 404 Not Found error
func (er *ExecutionResult) IsNotFoundError(err error) bool {
	return api.IsNotFound(err)
}

// Summary returns a summary message
func (er *ExecutionResult) Summary() string {
	successCount := 0
	failCount := 0

	for _, result := range er.results {
		if len(result.Errors) > 0 {
			failCount++
		} else {
			successCount++
		}
	}

	if failCount == 0 {
		return "Summary: all operations completed successfully"
	}

	return fmt.Sprintf("Summary: %d repositories succeeded, %d failed", 
		successCount, failCount)
}

// Err returns an error if any operations failed
func (er *ExecutionResult) Err() error {
	if !er.HasErrors() {
		return nil
	}
	// Return a simple error for exit code purposes
	return fmt.Errorf("some operations failed")
}
