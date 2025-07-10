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
package api

import (
	"context"
	"fmt"
	"github.com/google/go-github/v59/github"
	"github.com/tnagatomi/gh-fuda/option"
	"net/http"
	"strings"
)

// APIClient is a interface for the API client
type APIClient interface {
	CreateLabel(label option.Label, repo option.Repo) error
	DeleteLabel(label string, repo option.Repo) error
	ListLabels(repo option.Repo) ([]string, error)
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
		if strings.Contains(err.Error(), "already_exists") {
			return fmt.Errorf("label %q already exists for repository %q", label, repo)
		}
		return err
	}

	return nil
}

// DeleteLabel deletes a label in the repository
func (a *API) DeleteLabel(label string, repo option.Repo) error {
	_, err := a.client.Issues.DeleteLabel(context.Background(), repo.Owner, repo.Repo, label)

	if err != nil {
		return err
	}

	return nil
}

// ListLabels gets all label names in the repository
func (a *API) ListLabels(repo option.Repo) ([]string, error) {
	labels, _, err := a.client.Issues.ListLabels(context.Background(), repo.Owner, repo.Repo, nil)

	if err != nil {
		return nil, err
	}

	var labelNames []string
	for _, l := range labels {
		labelNames = append(labelNames, l.GetName())
	}

	return labelNames, nil
}
