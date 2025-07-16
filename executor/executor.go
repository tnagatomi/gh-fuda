/*
Copyright © 2024 Takayuki Nagatomi <tommyt6073@gmail.com>

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
	"io"
	"net/http"
	"strings"

	"github.com/tnagatomi/gh-fuda/api"
	"github.com/tnagatomi/gh-fuda/option"
	"github.com/tnagatomi/gh-fuda/parser"
)

// Executor composites github.Client and has dry-run option
type Executor struct {
	api    api.APIClient
	dryRun bool
}

// NewExecutor returns new Executor
func NewExecutor(client *http.Client, dryrun bool) (*Executor, error) {
	api, err := api.NewAPI(client)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize API client: %v", err)
	}

	return &Executor{
		api:    api,
		dryRun: dryrun,
	}, nil
}

// Create creates labels across multiple repositories
func (e *Executor) Create(out io.Writer, repoOption string, labelOption string) error {
	labels, err := parser.Label(labelOption)
	if err != nil {
		return fmt.Errorf("failed to parse label option: %v", err)
	}
	repos, err := parser.Repo(repoOption)
	if err != nil {
		return fmt.Errorf("failed to parse repo option: %v", err)
	}

	er := NewExecutionResult()

	for _, repo := range repos {
		repoResult := &RepoResult{
			Repo:   repo.String(),
			Errors: nil,
		}

		for _, label := range labels {
			if e.dryRun {
				_, _ = fmt.Fprintf(out, "Would create label %q for repository %q\n", label, repo)
				continue
			}

			err = e.api.CreateLabel(label, repo)
			if err != nil {
				repoResult.Errors = append(repoResult.Errors, err)
				_, _ = fmt.Fprintf(out, "Failed to create label %q for repository %q: %v\n", label, repo, err)
				continue
			}
			_, _ = fmt.Fprintf(out, "Created label %q for repository %q\n", label, repo)
		}

		if !e.dryRun {
			er.AddRepoResult(repoResult)
		}
	}

	if !e.dryRun {
		_, _ = fmt.Fprintf(out, "\n%s\n", er.Summary())
	}

	return er.Err()
}

// Delete deletes labels across multiple repositories
func (e *Executor) Delete(out io.Writer, repoOption string, labelOption string) error {
	labels := strings.Split(labelOption, ",")

	repos, err := parser.Repo(repoOption)
	if err != nil {
		return fmt.Errorf("failed to parse repo option: %v", err)
	}

	er := NewExecutionResult()

	for _, repo := range repos {
		repoResult := &RepoResult{
			Repo:   repo.String(),
			Errors: nil,
		}

		for _, label := range labels {
			if e.dryRun {
				_, _ = fmt.Fprintf(out, "Would delete label %q for repository %q\n", label, repo)
				continue
			}

			err = e.api.DeleteLabel(label, repo)
			if err != nil {
				repoResult.Errors = append(repoResult.Errors, err)
				_, _ = fmt.Fprintf(out, "Failed to delete label %q for repository %q: %v\n", label, repo, err)
				continue
			}
			_, _ = fmt.Fprintf(out, "Deleted label %q for repository %q\n", label, repo)
		}

		if !e.dryRun {
			er.AddRepoResult(repoResult)
		}
	}

	if !e.dryRun {
		_, _ = fmt.Fprintf(out, "\n%s\n", er.Summary())
	}

	return er.Err()
}

// Sync sync labels across multiple repositories
func (e *Executor) Sync(out io.Writer, repoOption string, labelOption string) error {
	repos, err := parser.Repo(repoOption)
	if err != nil {
		return fmt.Errorf("failed to parse repo option: %v", err)
	}

	labels, err := parser.Label(labelOption)
	if err != nil {
		return fmt.Errorf("failed to parse label option: %v", err)
	}

	er := NewExecutionResult()

	for _, repo := range repos {
		repoStr := repo.String()
		repoResult := &RepoResult{
			Repo:   repoStr,
			Errors: nil,
		}

		existingLabels, err := e.api.ListLabels(repo)
		if err != nil {
			repoResult.Errors = append(repoResult.Errors, err)
			_, _ = fmt.Fprintf(out, "Failed to list labels for repository %q: %v\n", repo, err)
			if !e.dryRun {
				er.AddRepoResult(repoResult)
			}
			continue
		}

		// Delete labels not in the new set
		for _, existing := range existingLabels {
			if labelNameExists(existing, labels) {
				continue
			}

			if e.dryRun {
				_, _ = fmt.Fprintf(out, "Would delete label %q for repository %q\n", existing, repo)
				continue
			}

			err = e.api.DeleteLabel(existing, repo)
			if err != nil {
				repoResult.Errors = append(repoResult.Errors, err)
				_, _ = fmt.Fprintf(out, "Failed to delete label %q for repository %q: %v\n", existing, repo, err)
			} else {
				_, _ = fmt.Fprintf(out, "Deleted label %q for repository %q\n", existing, repo)
			}
		}

		// Create or update labels
		for _, label := range labels {
			if stringExists(label.Name, existingLabels) {
				if e.dryRun {
					_, _ = fmt.Fprintf(out, "Would update label %q for repository %q\n", label, repo)
					continue
				}

				err = e.api.UpdateLabel(label, repo)
				if err != nil {
					repoResult.Errors = append(repoResult.Errors, err)
					_, _ = fmt.Fprintf(out, "Failed to update label %q for repository %q: %v\n", label, repo, err)
				} else {
					_, _ = fmt.Fprintf(out, "Updated label %q for repository %q\n", label, repo)
				}
			} else {
				// Create new label
				if e.dryRun {
					_, _ = fmt.Fprintf(out, "Would create label %q for repository %q\n", label, repo)
					continue
				}

				err = e.api.CreateLabel(label, repo)
				if err != nil {
					repoResult.Errors = append(repoResult.Errors, err)
					_, _ = fmt.Fprintf(out, "Failed to create label %q for repository %q: %v\n", label, repo, err)
				} else {
					_, _ = fmt.Fprintf(out, "Created label %q for repository %q\n", label, repo)
				}
			}
		}

		if !e.dryRun {
			er.AddRepoResult(repoResult)
		}
	}

	if !e.dryRun {
		_, _ = fmt.Fprintf(out, "\n%s\n", er.Summary())
	}

	return er.Err()
}

// Empty empties labels across multiple repositories
func (e *Executor) Empty(out io.Writer, repoOption string) error {
	repos, err := parser.Repo(repoOption)
	if err != nil {
		return fmt.Errorf("failed to parse repo option: %v", err)
	}

	er := NewExecutionResult()

	results := e.emptyLabels(out, repos)
	if !e.dryRun {
		for _, result := range results {
			er.AddRepoResult(result)
		}
		_, _ = fmt.Fprintf(out, "\n%s\n", er.Summary())
	}

	return er.Err()
}

func (e *Executor) emptyLabels(out io.Writer, repos []option.Repo) []*RepoResult {
	var results []*RepoResult

	for _, repo := range repos {
		repoResult := &RepoResult{
			Repo:   repo.String(),
			Errors: nil,
		}

		labels, err := e.api.ListLabels(repo)
		if err != nil {
			repoResult.Errors = append(repoResult.Errors, err)
			_, _ = fmt.Fprintf(out, "Failed to list labels for repository %q: %v\n", repo, err)
			results = append(results, repoResult)
			continue
		}

		for _, label := range labels {
			if e.dryRun {
				_, _ = fmt.Fprintf(out, "Would delete label %q for repository %q\n", label, repo)
				continue
			}

			err = e.api.DeleteLabel(label, repo)
			if err != nil {
				repoResult.Errors = append(repoResult.Errors, err)
				_, _ = fmt.Fprintf(out, "Failed to delete label %q for repository %q: %v\n", label, repo, err)
			} else {
				_, _ = fmt.Fprintf(out, "Deleted label %q for repository %q\n", label, repo)
			}
		}

		results = append(results, repoResult)
	}

	return results
}

// labelNameExists checks if a label name exists in a slice of labels
func labelNameExists(name string, labels []option.Label) bool {
	for _, label := range labels {
		if name == label.Name {
			return true
		}
	}
	return false
}

// stringExists checks if a string exists in a slice of strings
func stringExists(target string, strings []string) bool {
	for _, s := range strings {
		if target == s {
			return true
		}
	}
	return false
}
