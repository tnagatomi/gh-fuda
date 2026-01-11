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
	"sync"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/tnagatomi/gh-fuda/option"
)

// GraphQLAPI implements APIClient using GitHub GraphQL API
type GraphQLAPI struct {
	client *api.GraphQLClient
	// Cache for repository IDs to avoid redundant queries
	repoIDCache map[string]option.GraphQLID
	repoIDMu    sync.RWMutex
}

// NewGraphQLAPI creates a new GraphQL API client
func NewGraphQLAPI() (*GraphQLAPI, error) {
	client, err := api.NewGraphQLClient(api.ClientOptions{})
	if err != nil {
		return nil, err
	}
	return &GraphQLAPI{
		client:      client,
		repoIDCache: make(map[string]option.GraphQLID),
	}, nil
}

// GetRepositoryID fetches the GraphQL node ID for a repository
func (g *GraphQLAPI) GetRepositoryID(repo option.Repo) (option.GraphQLID, error) {
	cacheKey := repo.String()

	// Check cache with read lock
	g.repoIDMu.RLock()
	if id, ok := g.repoIDCache[cacheKey]; ok {
		g.repoIDMu.RUnlock()
		return id, nil
	}
	g.repoIDMu.RUnlock()

	var query struct {
		Repository struct {
			ID string
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]any{
		"owner": graphql.String(repo.Owner),
		"name":  graphql.String(repo.Repo),
	}

	err := g.client.Query("RepositoryID", &query, variables)
	if err != nil {
		return "", wrapGraphQLError(err, ResourceTypeRepository)
	}

	// Store in cache with write lock
	g.repoIDMu.Lock()
	g.repoIDCache[cacheKey] = option.GraphQLID(query.Repository.ID)
	g.repoIDMu.Unlock()

	return option.GraphQLID(query.Repository.ID), nil
}

// GetLabelID fetches the GraphQL node ID for a label in a repository
func (g *GraphQLAPI) GetLabelID(repo option.Repo, labelName string) (option.GraphQLID, error) {
	var query struct {
		Repository struct {
			Label struct {
				ID string
			} `graphql:"label(name: $labelName)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]any{
		"owner":     graphql.String(repo.Owner),
		"name":      graphql.String(repo.Repo),
		"labelName": graphql.String(labelName),
	}

	err := g.client.Query("LabelID", &query, variables)
	if err != nil {
		return "", wrapGraphQLError(err, ResourceTypeLabel)
	}

	if query.Repository.Label.ID == "" {
		return "", &NotFoundError{ResourceType: ResourceTypeLabel}
	}

	return option.GraphQLID(query.Repository.Label.ID), nil
}

// ListLabels fetches all labels in a repository with pagination
func (g *GraphQLAPI) ListLabels(repo option.Repo) ([]option.Label, error) {
	var allLabels []option.Label
	var cursor *graphql.String

	for {
		var query struct {
			Repository struct {
				Labels struct {
					Nodes []struct {
						Name        string
						Color       string
						Description string
					}
					PageInfo struct {
						HasNextPage bool
						EndCursor   string
					}
				} `graphql:"labels(first: 100, after: $cursor)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables := map[string]any{
			"owner":  graphql.String(repo.Owner),
			"name":   graphql.String(repo.Repo),
			"cursor": cursor,
		}

		err := g.client.Query("RepositoryLabels", &query, variables)
		if err != nil {
			return nil, wrapGraphQLError(err, ResourceTypeRepository)
		}

		for _, node := range query.Repository.Labels.Nodes {
			allLabels = append(allLabels, option.Label{
				Name:        node.Name,
				Color:       node.Color,
				Description: node.Description,
			})
		}

		if !query.Repository.Labels.PageInfo.HasNextPage {
			break
		}
		endCursor := graphql.String(query.Repository.Labels.PageInfo.EndCursor)
		cursor = &endCursor
	}

	return allLabels, nil
}

// CreateLabel creates a new label in a repository
func (g *GraphQLAPI) CreateLabel(label option.Label, repo option.Repo) error {
	repoID, err := g.GetRepositoryID(repo)
	if err != nil {
		return err
	}

	var mutation struct {
		CreateLabel struct {
			Label struct {
				ID string
			}
		} `graphql:"createLabel(input: $input)"`
	}

	type CreateLabelInput struct {
		RepositoryID string `json:"repositoryId"`
		Name         string `json:"name"`
		Color        string `json:"color"`
		Description  string `json:"description,omitempty"`
	}

	variables := map[string]any{
		"input": CreateLabelInput{
			RepositoryID: string(repoID),
			Name:         label.Name,
			Color:        label.Color,
			Description:  label.Description,
		},
	}

	err = g.client.Mutate("CreateLabel", &mutation, variables)
	if err != nil {
		return wrapGraphQLError(err, ResourceTypeLabel)
	}

	return nil
}

// UpdateLabel updates an existing label in a repository
func (g *GraphQLAPI) UpdateLabel(label option.Label, repo option.Repo) error {
	labelID, err := g.GetLabelID(repo, label.Name)
	if err != nil {
		return err
	}

	var mutation struct {
		UpdateLabel struct {
			Label struct {
				ID string
			}
		} `graphql:"updateLabel(input: $input)"`
	}

	type UpdateLabelInput struct {
		ID          string `json:"id"`
		Name        string `json:"name,omitempty"`
		Color       string `json:"color,omitempty"`
		Description string `json:"description,omitempty"`
	}

	variables := map[string]any{
		"input": UpdateLabelInput{
			ID:          string(labelID),
			Name:        label.Name,
			Color:       label.Color,
			Description: label.Description,
		},
	}

	err = g.client.Mutate("UpdateLabel", &mutation, variables)
	if err != nil {
		return wrapGraphQLError(err, ResourceTypeLabel)
	}

	return nil
}

// DeleteLabel deletes a label from a repository
func (g *GraphQLAPI) DeleteLabel(label string, repo option.Repo) error {
	labelID, err := g.GetLabelID(repo, label)
	if err != nil {
		return err
	}

	var mutation struct {
		DeleteLabel struct {
			ClientMutationID *string
		} `graphql:"deleteLabel(input: $input)"`
	}

	type DeleteLabelInput struct {
		ID string `json:"id"`
	}

	variables := map[string]any{
		"input": DeleteLabelInput{
			ID: string(labelID),
		},
	}

	err = g.client.Mutate("DeleteLabel", &mutation, variables)
	if err != nil {
		return wrapGraphQLError(err, ResourceTypeLabel)
	}

	return nil
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
	if strings.Contains(errMsg, "Could not resolve to a node") {
		return &NotFoundError{ResourceType: resourceType}
	}
	if strings.Contains(errMsg, "NOT_FOUND") {
		return &NotFoundError{ResourceType: resourceType}
	}
	if strings.Contains(errMsg, "FORBIDDEN") || strings.Contains(errMsg, "don't have permission") {
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

// escapeSearchQuery escapes special characters in a string for use in GitHub search queries.
// It escapes backslashes and double quotes to prevent query injection.
func escapeSearchQuery(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// SearchLabelables searches for issues, pull requests, and discussions with a specific label
func (g *GraphQLAPI) SearchLabelables(repo option.Repo, labelName string) ([]option.Labelable, error) {
	var allLabelables []option.Labelable

	issuesAndPRs, err := g.searchIssuesAndPRs(repo, labelName)
	if err != nil {
		return nil, err
	}
	allLabelables = append(allLabelables, issuesAndPRs...)

	discussions, err := g.searchDiscussions(repo, labelName)
	if err != nil {
		return nil, err
	}
	allLabelables = append(allLabelables, discussions...)

	return allLabelables, nil
}

// searchIssuesAndPRs searches for issues and pull requests with a specific label
func (g *GraphQLAPI) searchIssuesAndPRs(repo option.Repo, labelName string) ([]option.Labelable, error) {
	var allLabelables []option.Labelable
	var cursor *graphql.String

	searchQuery := fmt.Sprintf("repo:%s/%s label:\"%s\"", repo.Owner, repo.Repo, escapeSearchQuery(labelName))

	for {
		var query struct {
			Search struct {
				Nodes []struct {
					TypeName    string              `graphql:"__typename"`
					Issue       issueFragment       `graphql:"... on Issue"`
					PullRequest pullRequestFragment `graphql:"... on PullRequest"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"search(query: $query, type: ISSUE, first: 100, after: $cursor)"`
		}

		variables := map[string]any{
			"query":  graphql.String(searchQuery),
			"cursor": cursor,
		}

		err := g.client.Query("SearchIssuesAndPRs", &query, variables)
		if err != nil {
			return nil, wrapGraphQLError(err, ResourceTypeRepository)
		}

		for _, node := range query.Search.Nodes {
			switch node.TypeName {
			case "Issue":
				allLabelables = append(allLabelables, option.Labelable{
					ID:     option.GraphQLID(node.Issue.ID),
					Number: node.Issue.Number,
					Title:  node.Issue.Title,
					Type:   option.LabelableTypeIssue,
				})
			case "PullRequest":
				allLabelables = append(allLabelables, option.Labelable{
					ID:     option.GraphQLID(node.PullRequest.ID),
					Number: node.PullRequest.Number,
					Title:  node.PullRequest.Title,
					Type:   option.LabelableTypePullRequest,
				})
			}
		}

		if !query.Search.PageInfo.HasNextPage {
			break
		}
		endCursor := graphql.String(query.Search.PageInfo.EndCursor)
		cursor = &endCursor
	}

	return allLabelables, nil
}

type issueFragment struct {
	ID     string
	Number int
	Title  string
}

type pullRequestFragment struct {
	ID     string
	Number int
	Title  string
}

// searchDiscussions searches for discussions with a specific label in a repository
func (g *GraphQLAPI) searchDiscussions(repo option.Repo, labelName string) ([]option.Labelable, error) {
	var allLabelables []option.Labelable
	var cursor *graphql.String

	searchQuery := fmt.Sprintf("repo:%s/%s label:\"%s\"", repo.Owner, repo.Repo, escapeSearchQuery(labelName))

	for {
		var query struct {
			Search struct {
				Nodes []struct {
					Discussion discussionFragment `graphql:"... on Discussion"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"search(query: $query, type: DISCUSSION, first: 100, after: $cursor)"`
		}

		variables := map[string]any{
			"query":  graphql.String(searchQuery),
			"cursor": cursor,
		}

		err := g.client.Query("SearchDiscussions", &query, variables)
		if err != nil {
			errMsg := err.Error()
			// Check for explicit errors first
			if strings.Contains(errMsg, "Could not resolve to a Repository") {
				return nil, wrapGraphQLError(err, ResourceTypeRepository)
			}
			if strings.Contains(errMsg, "FORBIDDEN") || strings.Contains(errMsg, "don't have permission") {
				return nil, wrapGraphQLError(err, ResourceTypeRepository)
			}
			// If discussions are not enabled/supported, just return empty
			if strings.Contains(errMsg, "does not have discussions enabled") ||
				strings.Contains(errMsg, "Discussions are disabled") {
				return allLabelables, nil
			}
			return nil, wrapGraphQLError(err, ResourceTypeRepository)
		}

		for _, node := range query.Search.Nodes {
			if node.Discussion.ID != "" {
				allLabelables = append(allLabelables, option.Labelable{
					ID:     option.GraphQLID(node.Discussion.ID),
					Number: node.Discussion.Number,
					Title:  node.Discussion.Title,
					Type:   option.LabelableTypeDiscussion,
				})
			}
		}

		if !query.Search.PageInfo.HasNextPage {
			break
		}
		endCursor := graphql.String(query.Search.PageInfo.EndCursor)
		cursor = &endCursor
	}

	return allLabelables, nil
}

type discussionFragment struct {
	ID     string
	Number int
	Title  string
}

// AddLabelsToLabelable adds labels to a labelable resource (issue, PR, or discussion)
func (g *GraphQLAPI) AddLabelsToLabelable(labelableID option.GraphQLID, labelIDs []option.GraphQLID) error {
	var mutation struct {
		AddLabelsToLabelable struct {
			ClientMutationID *string
		} `graphql:"addLabelsToLabelable(input: $input)"`
	}

	type AddLabelsToLabelableInput struct {
		LabelableID string   `json:"labelableId"`
		LabelIDs    []string `json:"labelIds"`
	}

	// Convert GraphQLIDs to strings for the API
	stringLabelIDs := make([]string, len(labelIDs))
	for i, id := range labelIDs {
		stringLabelIDs[i] = string(id)
	}

	variables := map[string]any{
		"input": AddLabelsToLabelableInput{
			LabelableID: string(labelableID),
			LabelIDs:    stringLabelIDs,
		},
	}

	err := g.client.Mutate("AddLabelsToLabelable", &mutation, variables)
	if err != nil {
		return wrapGraphQLError(err, ResourceTypeLabel)
	}

	return nil
}

// RemoveLabelsFromLabelable removes labels from a labelable resource (issue, PR, or discussion)
func (g *GraphQLAPI) RemoveLabelsFromLabelable(labelableID option.GraphQLID, labelIDs []option.GraphQLID) error {
	var mutation struct {
		RemoveLabelsFromLabelable struct {
			ClientMutationID *string
		} `graphql:"removeLabelsFromLabelable(input: $input)"`
	}

	type RemoveLabelsFromLabelableInput struct {
		LabelableID string   `json:"labelableId"`
		LabelIDs    []string `json:"labelIds"`
	}

	// Convert GraphQLIDs to strings for the API
	stringLabelIDs := make([]string, len(labelIDs))
	for i, id := range labelIDs {
		stringLabelIDs[i] = string(id)
	}

	variables := map[string]any{
		"input": RemoveLabelsFromLabelableInput{
			LabelableID: string(labelableID),
			LabelIDs:    stringLabelIDs,
		},
	}

	err := g.client.Mutate("RemoveLabelsFromLabelable", &mutation, variables)
	if err != nil {
		return wrapGraphQLError(err, ResourceTypeLabel)
	}

	return nil
}
