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
	"io"
	"net/http"

	"github.com/tnagatomi/gh-fuda/api"
	"github.com/tnagatomi/gh-fuda/option"
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
func (e *Executor) Create(out io.Writer, repos []option.Repo, labels []option.Label) error {
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

			err := e.api.CreateLabel(label, repo)
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
func (e *Executor) Delete(out io.Writer, repos []option.Repo, labels []string) error {
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

			err := e.api.DeleteLabel(label, repo)
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
func (e *Executor) Sync(out io.Writer, repos []option.Repo, labels []option.Label) error {
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
			if labelNameExists(existing.Name, labels) {
				continue
			}

			if e.dryRun {
				_, _ = fmt.Fprintf(out, "Would delete label %q for repository %q\n", existing.Name, repo)
				continue
			}

			err = e.api.DeleteLabel(existing.Name, repo)
			if err != nil {
				repoResult.Errors = append(repoResult.Errors, err)
				_, _ = fmt.Fprintf(out, "Failed to delete label %q for repository %q: %v\n", existing.Name, repo, err)
			} else {
				_, _ = fmt.Fprintf(out, "Deleted label %q for repository %q\n", existing.Name, repo)
			}
		}

		// Create or update labels
		for _, label := range labels {
			if labelExists(label.Name, existingLabels) {
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

// List lists labels across multiple repositories
func (e *Executor) List(out io.Writer, repos []option.Repo) error {
	er := NewExecutionResult()

	for _, repo := range repos {
		repoResult := &RepoResult{
			Repo:   repo.String(),
			Errors: nil,
		}

		labels, err := e.api.ListLabels(repo)
		if err != nil {
			repoResult.Errors = append(repoResult.Errors, err)
			_, _ = fmt.Fprintf(out, "Failed to list labels for repository %q: %v\n", repo, err)
			er.AddRepoResult(repoResult)
			continue
		}

		if len(labels) == 0 {
			_, _ = fmt.Fprintf(out, "Repository %q has no labels\n", repo)
		} else {
			_, _ = fmt.Fprintf(out, "Labels for repository %q:\n", repo)
			for _, label := range labels {
				if label.Description == "" {
					_, _ = fmt.Fprintf(out, "  %s (#%s)\n", label.Name, label.Color)
				} else {
					_, _ = fmt.Fprintf(out, "  %s (#%s) - %s\n", label.Name, label.Color, label.Description)
				}
			}
		}

		er.AddRepoResult(repoResult)
	}

	_, _ = fmt.Fprintf(out, "\n%s\n", er.Summary())

	return er.Err()
}

// Empty empties labels across multiple repositories
func (e *Executor) Empty(out io.Writer, repos []option.Repo) error {
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
				_, _ = fmt.Fprintf(out, "Would delete label %q for repository %q\n", label.Name, repo)
				continue
			}

			err := e.api.DeleteLabel(label.Name, repo)
			if err != nil {
				repoResult.Errors = append(repoResult.Errors, err)
				_, _ = fmt.Fprintf(out, "Failed to delete label %q for repository %q: %v\n", label.Name, repo, err)
			} else {
				_, _ = fmt.Fprintf(out, "Deleted label %q for repository %q\n", label.Name, repo)
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

// labelExists checks if a label name exists in a slice of labels
func labelExists(name string, labels []option.Label) bool {
	for _, label := range labels {
		if name == label.Name {
			return true
		}
	}
	return false
}
