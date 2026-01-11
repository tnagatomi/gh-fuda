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
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/tnagatomi/gh-fuda/option"
)

// GraphQLAPI implements APIClient using GitHub GraphQL API
type GraphQLAPI struct {
	client *api.GraphQLClient
	// Cache for repository IDs to avoid redundant queries
	repoIDCache map[string]string
}

// NewGraphQLAPI creates a new GraphQL API client
func NewGraphQLAPI() (*GraphQLAPI, error) {
	client, err := api.NewGraphQLClient(api.ClientOptions{})
	if err != nil {
		return nil, err
	}
	return &GraphQLAPI{
		client:      client,
		repoIDCache: make(map[string]string),
	}, nil
}

// GetRepositoryID fetches the GraphQL node ID for a repository
func (g *GraphQLAPI) GetRepositoryID(repo option.Repo) (string, error) {
	cacheKey := repo.String()
	if id, ok := g.repoIDCache[cacheKey]; ok {
		return id, nil
	}

	var query struct {
		Repository struct {
			ID string
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]any{
		"owner": repo.Owner,
		"name":  repo.Repo,
	}

	err := g.client.Query("RepositoryID", &query, variables)
	if err != nil {
		return "", wrapGraphQLError(err, ResourceTypeRepository)
	}

	g.repoIDCache[cacheKey] = query.Repository.ID
	return query.Repository.ID, nil
}

// GetLabelID fetches the GraphQL node ID for a label in a repository
func (g *GraphQLAPI) GetLabelID(repo option.Repo, labelName string) (string, error) {
	var query struct {
		Repository struct {
			Label struct {
				ID string
			} `graphql:"label(name: $labelName)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]any{
		"owner":     repo.Owner,
		"name":      repo.Repo,
		"labelName": labelName,
	}

	err := g.client.Query("LabelID", &query, variables)
	if err != nil {
		return "", wrapGraphQLError(err, ResourceTypeLabel)
	}

	if query.Repository.Label.ID == "" {
		return "", &NotFoundError{ResourceType: ResourceTypeLabel}
	}

	return query.Repository.Label.ID, nil
}

// wrapGraphQLError converts GraphQL API errors to custom error types
func wrapGraphQLError(err error, resourceType ResourceType) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for common GraphQL error patterns
	if strings.Contains(errMsg, "Could not resolve to a Repository") {
		return &NotFoundError{ResourceType: ResourceTypeRepository}
	}
	if strings.Contains(errMsg, "Could not resolve to a Label") {
		return &NotFoundError{ResourceType: ResourceTypeLabel}
	}
	if strings.Contains(errMsg, "NOT_FOUND") {
		return &NotFoundError{ResourceType: resourceType}
	}
	if strings.Contains(errMsg, "FORBIDDEN") {
		return &ForbiddenError{}
	}
	if strings.Contains(errMsg, "UNAUTHORIZED") || strings.Contains(errMsg, "401") {
		return &UnauthorizedError{}
	}
	if strings.Contains(errMsg, "RATE_LIMITED") || strings.Contains(errMsg, "rate limit") {
		return &RateLimitError{}
	}
	if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "ALREADY_EXISTS") {
		return &AlreadyExistsError{ResourceType: resourceType}
	}
	if strings.Contains(errMsg, "INSUFFICIENT_SCOPES") {
		return &ScopeError{}
	}

	return fmt.Errorf("GraphQL API error: %s", errMsg)
}
