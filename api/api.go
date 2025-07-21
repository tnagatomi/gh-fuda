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
package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/go-github/v59/github"
	"github.com/tnagatomi/gh-fuda/option"
)

// APIClient is a interface for the API client
type APIClient interface {
	CreateLabel(label option.Label, repo option.Repo) error
	UpdateLabel(label option.Label, repo option.Repo) error
	DeleteLabel(label string, repo option.Repo) error
	ListLabels(repo option.Repo) ([]option.Label, error)
}

// API is a wrapper around the GitHub client
type API struct {
	client *github.Client
}

// NewAPI returns new API
func NewAPI(client *http.Client) (*API, error) {
	return &API{
		client: github.NewClient(client),
	}, nil
}

// CreateLabel creates a label in the repository
func (a *API) CreateLabel(label option.Label, repo option.Repo) error {
	githubLabel := &github.Label{
		Name:        github.String(label.Name),
		Description: github.String(label.Description),
		Color:       github.String(label.Color),
	}

	_, _, err := a.client.Issues.CreateLabel(context.Background(), repo.Owner, repo.Repo, githubLabel)

	if err != nil {
		var ghErr *github.ErrorResponse
		if errors.As(err, &ghErr) && ghErr.Response.StatusCode == 404 {
			return wrapGitHubError(err, ResourceTypeRepository)
		}
		return wrapGitHubError(err, ResourceTypeLabel)
	}

	return nil
}

// UpdateLabel updates an existing label in the repository
func (a *API) UpdateLabel(label option.Label, repo option.Repo) error {
	githubLabel := &github.Label{
		Name:        github.String(label.Name),
		Description: github.String(label.Description),
		Color:       github.String(label.Color),
	}

	_, _, err := a.client.Issues.EditLabel(context.Background(), repo.Owner, repo.Repo, label.Name, githubLabel)

	if err != nil {
		return wrapGitHubError(err, ResourceTypeLabel)
	}

	return nil
}

// DeleteLabel deletes a label in the repository
func (a *API) DeleteLabel(label string, repo option.Repo) error {
	_, err := a.client.Issues.DeleteLabel(context.Background(), repo.Owner, repo.Repo, label)

	if err != nil {
		return wrapGitHubError(err, ResourceTypeLabel)
	}

	return nil
}

// ListLabels gets all labels in the repository
func (a *API) ListLabels(repo option.Repo) ([]option.Label, error) {
	var allLabels []*github.Label
	
	opts := &github.ListOptions{
		PerPage: 100,
	}
	
	for {
		labels, resp, err := a.client.Issues.ListLabels(context.Background(), repo.Owner, repo.Repo, opts)
		if err != nil {
			return nil, wrapGitHubError(err, ResourceTypeRepository)
		}
		
		allLabels = append(allLabels, labels...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	var optionLabels []option.Label
	for _, l := range allLabels {
		optionLabels = append(optionLabels, option.Label{
			Name:        l.GetName(),
			Color:       l.GetColor(),
			Description: l.GetDescription(),
		})
	}

	return optionLabels, nil
}
