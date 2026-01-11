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
	"net/http"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/h2non/gock"
	"github.com/tnagatomi/gh-fuda/option"
)

func newTestGraphQLAPI(t *testing.T) *GraphQLAPI {
	t.Helper()
	client, err := api.NewGraphQLClient(api.ClientOptions{
		Host:      "github.com",
		AuthToken: "test-token",
	})
	if err != nil {
		t.Fatalf("failed to create GraphQL client: %v", err)
	}
	return &GraphQLAPI{
		client:      client,
		repoIDCache: make(map[string]option.GraphQLID),
	}
}

func TestGraphQLAPI_GetRepositoryID(t *testing.T) {
	tests := []struct {
		name       string
		repo       option.Repo
		mock       func()
		want       option.GraphQLID
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			repo: option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"id": "R_123456",
							},
						},
					})
			},
			want:    "R_123456",
			wantErr: false,
		},
		{
			name: "repository not found",
			repo: option.Repo{Owner: "owner", Repo: "nonexistent"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a Repository with the name 'nonexistent'.",
							},
						},
					})
			},
			want:       "",
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			got, err := g.GetRepositoryID(tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepositoryID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("GetRepositoryID() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if got != tt.want {
				t.Errorf("GetRepositoryID() = %v, want %v", got, tt.want)
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestGraphQLAPI_GetRepositoryID_Cache(t *testing.T) {
	defer gock.Off()

	// Track HTTP requests
	httpCallCount := 0
	gock.Observe(func(req *http.Request, mock gock.Mock) {
		httpCallCount++
	})
	defer gock.Observe(nil)

	// Set up mock for the first call only
	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"id": "R_cached",
				},
			},
		})

	g := newTestGraphQLAPI(t)
	repo := option.Repo{Owner: "owner", Repo: "repo"}

	// First call - should make HTTP request
	got1, err := g.GetRepositoryID(repo)
	if err != nil {
		t.Fatalf("first GetRepositoryID() error = %v", err)
	}
	if got1 != "R_cached" {
		t.Errorf("first GetRepositoryID() = %v, want R_cached", got1)
	}
	if httpCallCount != 1 {
		t.Errorf("after first call, httpCallCount = %d, want 1", httpCallCount)
	}

	// Second call - should use cache, no HTTP request
	got2, err := g.GetRepositoryID(repo)
	if err != nil {
		t.Fatalf("second GetRepositoryID() error = %v", err)
	}
	if got2 != "R_cached" {
		t.Errorf("second GetRepositoryID() = %v, want R_cached", got2)
	}
	if httpCallCount != 1 {
		t.Errorf("after second call, httpCallCount = %d, want 1 (cache should be used)", httpCallCount)
	}
}

func TestGraphQLAPI_GetLabelID(t *testing.T) {
	tests := []struct {
		name       string
		repo       option.Repo
		labelName  string
		mock       func()
		want       option.GraphQLID
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "success",
			repo:      option.Repo{Owner: "owner", Repo: "repo"},
			labelName: "bug",
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"label": map[string]any{
									"id": "LA_123456",
								},
							},
						},
					})
			},
			want:    "LA_123456",
			wantErr: false,
		},
		{
			name:      "label not found",
			repo:      option.Repo{Owner: "owner", Repo: "repo"},
			labelName: "nonexistent",
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"label": nil,
							},
						},
					})
			},
			want:       "",
			wantErr:    true,
			wantErrMsg: "label not found",
		},
		{
			name:      "repository not found",
			repo:      option.Repo{Owner: "owner", Repo: "nonexistent"},
			labelName: "bug",
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a Repository with the name 'nonexistent'.",
							},
						},
					})
			},
			want:       "",
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			got, err := g.GetLabelID(tt.repo, tt.labelName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLabelID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("GetLabelID() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if got != tt.want {
				t.Errorf("GetLabelID() = %v, want %v", got, tt.want)
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestGraphQLAPI_ListLabels(t *testing.T) {
	tests := []struct {
		name       string
		repo       option.Repo
		mock       func()
		want       []option.Label
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success with labels",
			repo: option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"labels": map[string]any{
									"nodes": []map[string]any{
										{"name": "bug", "color": "d73a4a", "description": "Something is broken"},
										{"name": "enhancement", "color": "a2eeef", "description": "New feature"},
									},
									"pageInfo": map[string]any{
										"hasNextPage": false,
										"endCursor":   "",
									},
								},
							},
						},
					})
			},
			want: []option.Label{
				{Name: "bug", Color: "d73a4a", Description: "Something is broken"},
				{Name: "enhancement", Color: "a2eeef", Description: "New feature"},
			},
			wantErr: false,
		},
		{
			name: "empty repository",
			repo: option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"labels": map[string]any{
									"nodes": []map[string]any{},
									"pageInfo": map[string]any{
										"hasNextPage": false,
										"endCursor":   "",
									},
								},
							},
						},
					})
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "repository not found",
			repo: option.Repo{Owner: "owner", Repo: "nonexistent"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a Repository with the name 'nonexistent'.",
							},
						},
					})
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			got, err := g.ListLabels(tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("ListLabels() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ListLabels() returned %d labels, want %d", len(got), len(tt.want))
				return
			}
			for i, label := range got {
				if label != tt.want[i] {
					t.Errorf("ListLabels()[%d] = %v, want %v", i, label, tt.want[i])
				}
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestGraphQLAPI_ListLabels_Pagination(t *testing.T) {
	defer gock.Off()

	// First page
	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"labels": map[string]any{
						"nodes": []map[string]any{
							{"name": "bug", "color": "d73a4a", "description": "Bug"},
						},
						"pageInfo": map[string]any{
							"hasNextPage": true,
							"endCursor":   "cursor1",
						},
					},
				},
			},
		})

	// Second page
	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"labels": map[string]any{
						"nodes": []map[string]any{
							{"name": "enhancement", "color": "a2eeef", "description": "Enhancement"},
						},
						"pageInfo": map[string]any{
							"hasNextPage": false,
							"endCursor":   "",
						},
					},
				},
			},
		})

	g := newTestGraphQLAPI(t)
	got, err := g.ListLabels(option.Repo{Owner: "owner", Repo: "repo"})
	if err != nil {
		t.Fatalf("ListLabels() error = %v", err)
	}

	want := []option.Label{
		{Name: "bug", Color: "d73a4a", Description: "Bug"},
		{Name: "enhancement", Color: "a2eeef", Description: "Enhancement"},
	}

	if len(got) != len(want) {
		t.Errorf("ListLabels() returned %d labels, want %d", len(got), len(want))
		return
	}
	for i, label := range got {
		if label != want[i] {
			t.Errorf("ListLabels()[%d] = %v, want %v", i, label, want[i])
		}
	}

	if !gock.IsDone() {
		t.Errorf("pending mocks: %d", len(gock.Pending()))
	}
}

func TestGraphQLAPI_CreateLabel(t *testing.T) {
	tests := []struct {
		name       string
		label      option.Label
		repo       option.Repo
		mock       func()
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:  "success",
			label: option.Label{Name: "bug", Color: "d73a4a", Description: "Something is broken"},
			repo:  option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				// GetRepositoryID call
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"id": "R_123456",
							},
						},
					})
				// CreateLabel mutation
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"createLabel": map[string]any{
								"label": map[string]any{
									"id": "LA_123456",
								},
							},
						},
					})
			},
			wantErr: false,
		},
		{
			name:  "repository not found",
			label: option.Label{Name: "bug", Color: "d73a4a", Description: "Something is broken"},
			repo:  option.Repo{Owner: "owner", Repo: "nonexistent"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a Repository with the name 'nonexistent'.",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
		{
			name:  "label already exists",
			label: option.Label{Name: "bug", Color: "d73a4a", Description: "Something is broken"},
			repo:  option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				// GetRepositoryID call
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"id": "R_123456",
							},
						},
					})
				// CreateLabel mutation - label already exists
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"createLabel": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "UNPROCESSABLE",
								"message": "Label already exists",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "label already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			err := g.CreateLabel(tt.label, tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("CreateLabel() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestGraphQLAPI_UpdateLabel(t *testing.T) {
	tests := []struct {
		name       string
		label      option.Label
		repo       option.Repo
		mock       func()
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:  "success",
			label: option.Label{Name: "bug", Color: "ff0000", Description: "Updated description"},
			repo:  option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				// GetLabelID call
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"label": map[string]any{
									"id": "LA_123456",
								},
							},
						},
					})
				// UpdateLabel mutation
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"updateLabel": map[string]any{
								"label": map[string]any{
									"id": "LA_123456",
								},
							},
						},
					})
			},
			wantErr: false,
		},
		{
			name:  "label not found",
			label: option.Label{Name: "nonexistent", Color: "ff0000", Description: "Description"},
			repo:  option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"label": nil,
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "label not found",
		},
		{
			name:  "repository not found",
			label: option.Label{Name: "bug", Color: "ff0000", Description: "Description"},
			repo:  option.Repo{Owner: "owner", Repo: "nonexistent"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a Repository with the name 'nonexistent'.",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			err := g.UpdateLabel(tt.label, tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("UpdateLabel() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestGraphQLAPI_DeleteLabel(t *testing.T) {
	tests := []struct {
		name       string
		label      string
		repo       option.Repo
		mock       func()
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:  "success",
			label: "bug",
			repo:  option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				// GetLabelID call
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"label": map[string]any{
									"id": "LA_123456",
								},
							},
						},
					})
				// DeleteLabel mutation
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"deleteLabel": map[string]any{
								"clientMutationId": nil,
							},
						},
					})
			},
			wantErr: false,
		},
		{
			name:  "label not found",
			label: "nonexistent",
			repo:  option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"label": nil,
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "label not found",
		},
		{
			name:  "repository not found",
			label: "bug",
			repo:  option.Repo{Owner: "owner", Repo: "nonexistent"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a Repository with the name 'nonexistent'.",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
		{
			name:  "forbidden",
			label: "bug",
			repo:  option.Repo{Owner: "owner", Repo: "repo"},
			mock: func() {
				// GetLabelID call
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"label": map[string]any{
									"id": "LA_123456",
								},
							},
						},
					})
				// DeleteLabel mutation - forbidden
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"deleteLabel": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "FORBIDDEN",
								"message": "You don't have permission to delete this label",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			err := g.DeleteLabel(tt.label, tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("DeleteLabel() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestGraphQLAPI_SearchLabelables(t *testing.T) {
	tests := []struct {
		name       string
		repo       option.Repo
		labelName  string
		mock       func()
		want       []option.Labelable
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "success - issues and PRs",
			repo:      option.Repo{Owner: "owner", Repo: "repo"},
			labelName: "bug",
			mock: func() {
				// Issue/PR search returns results
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": map[string]any{
								"nodes": []map[string]any{
									{
										"__typename": "Issue",
										"id":         "I_123",
										"number":     1,
										"title":      "Bug issue",
									},
									{
										"__typename": "PullRequest",
										"id":         "PR_456",
										"number":     2,
										"title":      "Fix bug PR",
									},
								},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					})
				// Discussion search returns empty
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": map[string]any{
								"nodes":    []map[string]any{},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					})
			},
			want: []option.Labelable{
				{ID: "I_123", Number: 1, Title: "Bug issue", Type: option.LabelableTypeIssue},
				{ID: "PR_456", Number: 2, Title: "Fix bug PR", Type: option.LabelableTypePullRequest},
			},
			wantErr: false,
		},
		{
			name:      "success - with discussions",
			repo:      option.Repo{Owner: "owner", Repo: "repo"},
			labelName: "question",
			mock: func() {
				// Issue/PR search returns empty
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": map[string]any{
								"nodes": []map[string]any{},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					})
				// Discussion search returns result
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": map[string]any{
								"nodes": []map[string]any{
									{
										"id":     "D_789",
										"number": 3,
										"title":  "Question discussion",
									},
								},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					})
			},
			want: []option.Labelable{
				{ID: "D_789", Number: 3, Title: "Question discussion", Type: option.LabelableTypeDiscussion},
			},
			wantErr: false,
		},
		{
			name:      "empty result",
			repo:      option.Repo{Owner: "owner", Repo: "repo"},
			labelName: "nonexistent",
			mock: func() {
				// Issue/PR search returns empty
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": map[string]any{
								"nodes": []map[string]any{},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					})
				// Discussion search returns empty
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": map[string]any{
								"nodes":    []map[string]any{},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					})
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:      "issue search - repository not found",
			repo:      option.Repo{Owner: "owner", Repo: "nonexistent"},
			labelName: "bug",
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a Repository with the name 'nonexistent'.",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
		{
			name:      "issue search - forbidden",
			repo:      option.Repo{Owner: "owner", Repo: "private"},
			labelName: "bug",
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "FORBIDDEN",
								"message": "You don't have permission to access this repository",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "forbidden",
		},
		{
			name:      "discussion search - repository not found",
			repo:      option.Repo{Owner: "owner", Repo: "nonexistent"},
			labelName: "bug",
			mock: func() {
				// Issue search succeeds
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": map[string]any{
								"nodes": []map[string]any{},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					})
				// Discussion search fails with repository not found
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a Repository with the name 'nonexistent'.",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
		{
			name:      "discussion search - forbidden",
			repo:      option.Repo{Owner: "owner", Repo: "private"},
			labelName: "bug",
			mock: func() {
				// Issue search succeeds
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": map[string]any{
								"nodes": []map[string]any{
									{
										"__typename": "Issue",
										"id":         "I_123",
										"number":     1,
										"title":      "Bug issue",
									},
								},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					})
				// Discussion search fails with forbidden
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"search": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "FORBIDDEN",
								"message": "You don't have permission to access discussions",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			got, err := g.SearchLabelables(tt.repo, tt.labelName)

			if (err != nil) != tt.wantErr {
				t.Errorf("SearchLabelables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("SearchLabelables() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("SearchLabelables() got %d items, want %d", len(got), len(tt.want))
					return
				}
				for i, item := range got {
					if item != tt.want[i] {
						t.Errorf("SearchLabelables()[%d] = %v, want %v", i, item, tt.want[i])
					}
				}
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestGraphQLAPI_AddLabelsToLabelable(t *testing.T) {
	tests := []struct {
		name        string
		labelableID option.GraphQLID
		labelIDs    []option.GraphQLID
		mock        func()
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "success - single label",
			labelableID: "I_123",
			labelIDs:    []option.GraphQLID{"LA_456"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"addLabelsToLabelable": map[string]any{
								"clientMutationId": nil,
							},
						},
					})
			},
			wantErr: false,
		},
		{
			name:        "success - multiple labels",
			labelableID: "PR_123",
			labelIDs:    []option.GraphQLID{"LA_456", "LA_789"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"addLabelsToLabelable": map[string]any{
								"clientMutationId": nil,
							},
						},
					})
			},
			wantErr: false,
		},
		{
			name:        "labelable not found",
			labelableID: "I_nonexistent",
			labelIDs:    []option.GraphQLID{"LA_456"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"addLabelsToLabelable": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a node with the global id of 'I_nonexistent'",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "label not found",
		},
		{
			name:        "forbidden",
			labelableID: "I_123",
			labelIDs:    []option.GraphQLID{"LA_456"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"addLabelsToLabelable": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "FORBIDDEN",
								"message": "You don't have permission to add labels",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			err := g.AddLabelsToLabelable(tt.labelableID, tt.labelIDs)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddLabelsToLabelable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("AddLabelsToLabelable() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestGraphQLAPI_RemoveLabelsFromLabelable(t *testing.T) {
	tests := []struct {
		name        string
		labelableID option.GraphQLID
		labelIDs    []option.GraphQLID
		mock        func()
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "success - single label",
			labelableID: "I_123",
			labelIDs:    []option.GraphQLID{"LA_456"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"removeLabelsFromLabelable": map[string]any{
								"clientMutationId": nil,
							},
						},
					})
			},
			wantErr: false,
		},
		{
			name:        "success - multiple labels",
			labelableID: "PR_123",
			labelIDs:    []option.GraphQLID{"LA_456", "LA_789"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"removeLabelsFromLabelable": map[string]any{
								"clientMutationId": nil,
							},
						},
					})
			},
			wantErr: false,
		},
		{
			name:        "labelable not found",
			labelableID: "I_nonexistent",
			labelIDs:    []option.GraphQLID{"LA_456"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"removeLabelsFromLabelable": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "NOT_FOUND",
								"message": "Could not resolve to a node with the global id of 'I_nonexistent'",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "label not found",
		},
		{
			name:        "forbidden",
			labelableID: "I_123",
			labelIDs:    []option.GraphQLID{"LA_456"},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(map[string]any{
						"data": map[string]any{
							"removeLabelsFromLabelable": nil,
						},
						"errors": []map[string]any{
							{
								"type":    "FORBIDDEN",
								"message": "You don't have permission to remove labels",
							},
						},
					})
			},
			wantErr:    true,
			wantErrMsg: "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			g := newTestGraphQLAPI(t)
			err := g.RemoveLabelsFromLabelable(tt.labelableID, tt.labelIDs)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveLabelsFromLabelable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("RemoveLabelsFromLabelable() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestEscapeSearchQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special characters",
			input: "bug",
			want:  "bug",
		},
		{
			name:  "double quotes",
			input: `label with "quotes"`,
			want:  `label with \"quotes\"`,
		},
		{
			name:  "backslash",
			input: `label\with\backslashes`,
			want:  `label\\with\\backslashes`,
		},
		{
			name:  "both quotes and backslashes",
			input: `label "with" \both`,
			want:  `label \"with\" \\both`,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeSearchQuery(tt.input)
			if got != tt.want {
				t.Errorf("escapeSearchQuery(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
