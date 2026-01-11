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
		Host: "github.com",
	})
	if err != nil {
		t.Fatalf("failed to create GraphQL client: %v", err)
	}
	return &GraphQLAPI{
		client:      client,
		repoIDCache: make(map[string]string),
	}
}

func TestGraphQLAPI_GetRepositoryID(t *testing.T) {
	tests := []struct {
		name       string
		repo       option.Repo
		mock       func()
		want       string
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
		want       string
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
