package api

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/h2non/gock"
	"github.com/tnagatomi/gh-fuda/option"
)

func TestCreateLabel(t *testing.T) {
	type args struct {
		label option.Label
		repo  option.Repo
	}
	tests := []struct {
		name       string
		args       args
		mock       func()
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				label: option.Label{
					Name:        "bug",
					Description: "This is a bug",
					Color:       "ff0000",
				},
				repo: option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(201).
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"})
			},
			wantErr:    false,
		},
		{
			name: "unauthorized",
			args: args{
				label: option.Label{
					Name:        "bug",
					Description: "This is a bug",
					Color:       "ff0000",
				},
				repo: option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(401).
					JSON(map[string]string{"message": "Bad credentials"})
			},
			wantErr:    true,
			wantErrMsg: "unauthorized",
		},
		{
			name: "forbidden",
			args: args{
				label: option.Label{
					Name:        "bug",
					Description: "This is a bug",
					Color:       "ff0000",
				},
				repo: option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(403).
					JSON(map[string]string{"message": "Resource not accessible by integration"})
			},
			wantErr:    true,
			wantErrMsg: "forbidden",
		},
		{
			name: "rate limit exceeded",
			args: args{
				label: option.Label{
					Name:        "bug",
					Description: "This is a bug",
					Color:       "ff0000",
				},
				repo: option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(429).
					JSON(map[string]string{"message": "API rate limit exceeded"})
			},
			wantErr:    true,
			wantErrMsg: "rate limit exceeded",
		},
		{
			name: "repository not found",
			args: args{
				label: option.Label{
					Name:        "bug",
					Description: "This is a bug",
					Color:       "ff0000",
				},
				repo: option.Repo{
					Owner: "tnagatomi",
					Repo:  "non-existent-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/non-existent-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(404).
					JSON(map[string]string{"message": "Not Found"})
			},
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
		{
			name: "already exists",
			args: args{
				label: option.Label{
					Name:        "bug",
					Description: "This is a bug",
					Color:       "ff0000",
				},
				repo: option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(422).
					JSON(map[string]any{
						"errors": []map[string]string{
							{
								"code": "already_exists",
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

			api, err := NewAPI(http.DefaultClient)
			if err != nil {
				t.Fatalf("NewAPI() error = %v", err)
			}

			err = api.CreateLabel(tt.args.label, tt.args.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLabel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("CreateLabel() error = %v, wantErrMsg = %v", err.Error(), tt.wantErrMsg)
			}
			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestDeleteLabel(t *testing.T) {
	type args struct {
		label string
		repo  option.Repo
	}
	tests := []struct {
		name       string
		args       args
		mock       func()
		wantErr	   bool
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				label: "bug",
				repo:  option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/bug").
					Reply(204)
			},
			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				label: "bug",
				repo:  option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/bug").
					Reply(404).
					JSON(map[string]string{"message": "Not Found"})
			},
			wantErr: true,
			wantErrMsg: "label not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			api, err := NewAPI(http.DefaultClient)
			if err != nil {
				t.Fatalf("NewAPI() error = %v", err)
			}

			err = api.DeleteLabel(tt.args.label, tt.args.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteLabel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("DeleteLabel() error = %v, wantErrMsg = %v", err.Error(), tt.wantErrMsg)
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestUpdateLabel(t *testing.T) {
	type args struct {
		label option.Label
		repo  option.Repo
	}
	tests := []struct {
		name       string
		args       args
		mock       func()
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				label: option.Label{
					Name:        "bug",
					Description: "Updated bug description",
					Color:       "d73a4a",
				},
				repo: option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Patch("/repos/tnagatomi/mock-repo/labels/bug").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "Updated bug description", "color": "d73a4a"}).
					Reply(200).
					JSON(map[string]string{"name": "bug", "description": "Updated bug description", "color": "d73a4a"})
			},
			wantErr: false,
		},
		{
			name: "label not found",
			args: args{
				label: option.Label{
					Name:        "nonexistent",
					Description: "This label does not exist",
					Color:       "ff0000",
				},
				repo: option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Patch("/repos/tnagatomi/mock-repo/labels/nonexistent").
					Reply(404).
					JSON(map[string]string{"message": "Not Found"})
			},
			wantErr:    true,
			wantErrMsg: "label not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			api, err := NewAPI(http.DefaultClient)
			if err != nil {
				t.Fatalf("NewAPI() error = %v", err)
			}

			err = api.UpdateLabel(tt.args.label, tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateLabel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("UpdateLabel() error = %v, wantErrMsg = %v", err.Error(), tt.wantErrMsg)
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestListLabels(t *testing.T) {
	type args struct {
		repo  option.Repo
	}
	tests := []struct {
		name       string
		args       args
		mock       func()
		want       []option.Label
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				repo:  option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/mock-repo/labels").
					MatchParam("per_page", "100").
					Reply(200).
					JSON([]map[string]string{{"name": "bug", "description": "This is a bug", "color": "ff0000"}, {"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}, {"name": "question", "description": "This is a question", "color": "0000ff"}})
			},
			want: []option.Label{
				{Name: "bug", Color: "ff0000", Description: "This is a bug"},
				{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
				{Name: "question", Color: "0000ff", Description: "This is a question"},
			},
			wantErr: false,
		},
		{
			name: "no labels",
			args: args{
				repo:  option.Repo{
					Owner: "tnagatomi",
					Repo:  "mock-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/mock-repo/labels").
					MatchParam("per_page", "100").
					Reply(200).
					JSON([]map[string]string{})
			},
			want: []option.Label{},
			wantErr: false,
		},
		{
			name: "repository not found",
			args: args{
				repo:  option.Repo{
					Owner: "tnagatomi",
					Repo:  "non-existent-repo",
				},
			},
			mock: func() {
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/non-existent-repo/labels").
					MatchParam("per_page", "100").
					Reply(404).
					JSON(map[string]string{"message": "Not Found"})
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "repository not found",
		},
		{
			name: "pagination - multiple pages",
			args: args{
				repo:  option.Repo{
					Owner: "tnagatomi",
					Repo:  "many-labels-repo",
				},
			},
			mock: func() {
				// First page
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/many-labels-repo/labels").
					MatchParam("per_page", "100").
					Reply(200).
					SetHeader("Link", `<https://api.github.com/repos/tnagatomi/many-labels-repo/labels?per_page=100&page=2>; rel="next", <https://api.github.com/repos/tnagatomi/many-labels-repo/labels?per_page=100&page=2>; rel="last"`).
					JSON([]map[string]string{
						{"name": "label1", "description": "Description 1", "color": "ff0000"},
						{"name": "label2", "description": "Description 2", "color": "00ff00"},
						{"name": "label3", "description": "Description 3", "color": "0000ff"},
					})
				
				// Second page (last page)
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/many-labels-repo/labels").
					MatchParam("per_page", "100").
					MatchParam("page", "2").
					Reply(200).
					JSON([]map[string]string{
						{"name": "label4", "description": "Description 4", "color": "ffff00"},
						{"name": "label5", "description": "Description 5", "color": "ff00ff"},
					})
			},
			want: []option.Label{
				{Name: "label1", Color: "ff0000", Description: "Description 1"},
				{Name: "label2", Color: "00ff00", Description: "Description 2"},
				{Name: "label3", Color: "0000ff", Description: "Description 3"},
				{Name: "label4", Color: "ffff00", Description: "Description 4"},
				{Name: "label5", Color: "ff00ff", Description: "Description 5"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mock != nil {
				tt.mock()
			}

			api, err := NewAPI(http.DefaultClient)
			if err != nil {
				t.Fatalf("NewAPI() error = %v", err)
			}

			labels, err := api.ListLabels(tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListLabels() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("ListLabels() error = %v, wantErrMsg = %v", err.Error(), tt.wantErrMsg)
			}

			if !cmp.Equal(labels, tt.want, cmpopts.EquateEmpty()) {
				t.Errorf("ListLabels() = %v, want %v", labels, tt.want)
			}

			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}
