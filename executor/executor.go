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

	"github.com/tnagatomi/gh-fuda/api"
	"github.com/tnagatomi/gh-fuda/option"
)

// Executor composites github.Client and has dry-run option
type Executor struct {
	api    api.APIClient
	dryRun bool
}

// NewExecutor returns new Executor
func NewExecutor(dryrun bool) (*Executor, error) {
	apiClient, err := api.NewGraphQLAPI()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize API client: %v", err)
	}

	return &Executor{
		api:    apiClient,
		dryRun: dryrun,
	}, nil
}

// Create creates labels across multiple repositories
// If force is true, updates existing labels instead of failing
func (e *Executor) Create(out io.Writer, repos []option.Repo, labels []option.Label, force bool) error {
	// Dry-run mode: execute sequentially with immediate output
	if e.dryRun {
		return e.createDryRun(out, repos, labels, force)
	}

	// Normal mode: execute in parallel
	return e.createParallel(out, repos, labels, force)
}

func (e *Executor) createDryRun(out io.Writer, repos []option.Repo, labels []option.Label, force bool) error {
	var hasError bool
	for _, repo := range repos {
		var existingLabels []option.Label
		if force {
			var err error
			existingLabels, err = e.api.ListLabels(repo)
			if err != nil {
				_, _ = fmt.Fprintf(out, "Failed to list labels for repository %q: %v\n", repo, err)
				hasError = true
				continue
			}
		}

		for _, label := range labels {
			if force && labelExists(label.Name, existingLabels) {
				_, _ = fmt.Fprintf(out, "Would update label %q for repository %q\n", label, repo)
			} else {
				_, _ = fmt.Fprintf(out, "Would create label %q for repository %q\n", label, repo)
			}
		}
	}
	if hasError {
		return fmt.Errorf("some operations failed")
	}
	return nil
}

func (e *Executor) createParallel(out io.Writer, repos []option.Repo, labels []option.Label, force bool) error {
	wp := NewWorkerPool(out)
	jobs := make([]Job, len(repos))

	for i, repo := range repos {
		repo := repo // capture loop variable
		jobs[i] = Job{
			ID: i,
			Func: func() *JobResult {
				return e.createLabelsForRepo(repo, labels, force)
			},
		}
	}

	results := wp.Run(jobs)
	wp.ClearProgress()

	// Output all results together
	er := NewExecutionResult()
	for i, result := range results {
		_, _ = fmt.Fprint(out, result.Output)
		er.AddRepoResult(&RepoResult{
			Repo:   repos[i].String(),
			Errors: result.Errors,
		})
	}

	_, _ = fmt.Fprintf(out, "\n%s\n", er.Summary())
	return er.Err()
}

func (e *Executor) createLabelsForRepo(repo option.Repo, labels []option.Label, force bool) *JobResult {
	var output string
	var errors []error

	for _, label := range labels {
		err := e.api.CreateLabel(label, repo)
		if err != nil {
			// If force flag is set and label already exists, try to update it
			if force && api.IsAlreadyExists(err) {
				err = e.api.UpdateLabel(label, repo)
				if err != nil {
					output += fmt.Sprintf("Failed to update label %q for repository %q: %v\n", label, repo, err)
					errors = append(errors, err)
					continue
				}
				output += fmt.Sprintf("Updated label %q for repository %q\n", label, repo)
				continue
			}

			output += fmt.Sprintf("Failed to create label %q for repository %q: %v\n", label, repo, err)
			errors = append(errors, err)
			continue
		}
		output += fmt.Sprintf("Created label %q for repository %q\n", label, repo)
	}

	return &JobResult{
		Output:  output,
		Success: len(errors) == 0,
		Errors:  errors,
	}
}

// Delete deletes labels across multiple repositories
func (e *Executor) Delete(out io.Writer, repos []option.Repo, labels []string) error {
	// Dry-run mode: execute sequentially with immediate output
	if e.dryRun {
		return e.deleteDryRun(out, repos, labels)
	}

	// Normal mode: execute in parallel
	return e.deleteParallel(out, repos, labels)
}

func (e *Executor) deleteDryRun(out io.Writer, repos []option.Repo, labels []string) error {
	for _, repo := range repos {
		for _, label := range labels {
			_, _ = fmt.Fprintf(out, "Would delete label %q for repository %q\n", label, repo)
		}
	}
	return nil
}

func (e *Executor) deleteParallel(out io.Writer, repos []option.Repo, labels []string) error {
	wp := NewWorkerPool(out)
	jobs := make([]Job, len(repos))

	for i, repo := range repos {
		repo := repo // capture loop variable
		jobs[i] = Job{
			ID: i,
			Func: func() *JobResult {
				return e.deleteLabelsForRepo(repo, labels)
			},
		}
	}

	results := wp.Run(jobs)
	wp.ClearProgress()

	// Output all results together
	er := NewExecutionResult()
	for i, result := range results {
		_, _ = fmt.Fprint(out, result.Output)
		er.AddRepoResult(&RepoResult{
			Repo:   repos[i].String(),
			Errors: result.Errors,
		})
	}

	_, _ = fmt.Fprintf(out, "\n%s\n", er.Summary())
	return er.Err()
}

func (e *Executor) deleteLabelsForRepo(repo option.Repo, labels []string) *JobResult {
	var output string
	var errors []error

	for _, label := range labels {
		err := e.api.DeleteLabel(label, repo)
		if err != nil {
			output += fmt.Sprintf("Failed to delete label %q for repository %q: %v\n", label, repo, err)
			errors = append(errors, err)
			continue
		}
		output += fmt.Sprintf("Deleted label %q for repository %q\n", label, repo)
	}

	return &JobResult{
		Output:  output,
		Success: len(errors) == 0,
		Errors:  errors,
	}
}

// Sync sync labels across multiple repositories
func (e *Executor) Sync(out io.Writer, repos []option.Repo, labels []option.Label) error {
	// Dry-run mode: execute sequentially with immediate output
	if e.dryRun {
		return e.syncDryRun(out, repos, labels)
	}

	// Normal mode: execute in parallel
	return e.syncParallel(out, repos, labels)
}

func (e *Executor) syncDryRun(out io.Writer, repos []option.Repo, labels []option.Label) error {
	var hasError bool
	for _, repo := range repos {
		existingLabels, err := e.api.ListLabels(repo)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Failed to list labels for repository %q: %v\n", repo, err)
			hasError = true
			continue
		}

		// Delete labels not in the new set
		for _, existing := range existingLabels {
			if labelExists(existing.Name, labels) {
				continue
			}
			_, _ = fmt.Fprintf(out, "Would delete label %q for repository %q\n", existing.Name, repo)
		}

		// Create or update labels
		for _, label := range labels {
			if labelExists(label.Name, existingLabels) {
				_, _ = fmt.Fprintf(out, "Would update label %q for repository %q\n", label, repo)
			} else {
				_, _ = fmt.Fprintf(out, "Would create label %q for repository %q\n", label, repo)
			}
		}
	}
	if hasError {
		return fmt.Errorf("some operations failed")
	}
	return nil
}

func (e *Executor) syncParallel(out io.Writer, repos []option.Repo, labels []option.Label) error {
	wp := NewWorkerPool(out)
	jobs := make([]Job, len(repos))

	for i, repo := range repos {
		repo := repo // capture loop variable
		jobs[i] = Job{
			ID: i,
			Func: func() *JobResult {
				return e.syncLabelsForRepo(repo, labels)
			},
		}
	}

	results := wp.Run(jobs)
	wp.ClearProgress()

	// Output all results together
	er := NewExecutionResult()
	for i, result := range results {
		_, _ = fmt.Fprint(out, result.Output)
		er.AddRepoResult(&RepoResult{
			Repo:   repos[i].String(),
			Errors: result.Errors,
		})
	}

	_, _ = fmt.Fprintf(out, "\n%s\n", er.Summary())
	return er.Err()
}

func (e *Executor) syncLabelsForRepo(repo option.Repo, labels []option.Label) *JobResult {
	var output string
	var errors []error

	existingLabels, err := e.api.ListLabels(repo)
	if err != nil {
		output += fmt.Sprintf("Failed to list labels for repository %q: %v\n", repo, err)
		errors = append(errors, err)
		return &JobResult{
			Output:  output,
			Success: false,
			Errors:  errors,
		}
	}

	// Delete labels not in the new set
	for _, existing := range existingLabels {
		if labelExists(existing.Name, labels) {
			continue
		}

		err = e.api.DeleteLabel(existing.Name, repo)
		if err != nil {
			output += fmt.Sprintf("Failed to delete label %q for repository %q: %v\n", existing.Name, repo, err)
			errors = append(errors, err)
		} else {
			output += fmt.Sprintf("Deleted label %q for repository %q\n", existing.Name, repo)
		}
	}

	// Create or update labels
	for _, label := range labels {
		if labelExists(label.Name, existingLabels) {
			err = e.api.UpdateLabel(label, repo)
			if err != nil {
				output += fmt.Sprintf("Failed to update label %q for repository %q: %v\n", label, repo, err)
				errors = append(errors, err)
			} else {
				output += fmt.Sprintf("Updated label %q for repository %q\n", label, repo)
			}
		} else {
			err = e.api.CreateLabel(label, repo)
			if err != nil {
				output += fmt.Sprintf("Failed to create label %q for repository %q: %v\n", label, repo, err)
				errors = append(errors, err)
			} else {
				output += fmt.Sprintf("Created label %q for repository %q\n", label, repo)
			}
		}
	}

	return &JobResult{
		Output:  output,
		Success: len(errors) == 0,
		Errors:  errors,
	}
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

// labelExists checks if a label name exists in a slice of labels
func labelExists(name string, labels []option.Label) bool {
	for _, label := range labels {
		if name == label.Name {
			return true
		}
	}
	return false
}
