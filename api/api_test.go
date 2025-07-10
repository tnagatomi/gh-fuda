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
			wantErrMsg: "",
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
			wantErrMsg: "label \"bug\" already exists for repository \"tnagatomi/mock-repo\"",
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

			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("CreateLabel() error = %v, wantErr = %v", err, tt.wantErrMsg)
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
		wantErr	 bool
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
					Reply(404)
			},
			wantErr: true,
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
				t.Errorf("DeleteLabel() error = %v", err)
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
		want []string
		wantErr bool
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
					Reply(200).
					JSON([]map[string]string{{"name": "bug", "description": "This is a bug", "color": "ff0000"}, {"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}, {"name": "question", "description": "This is a question", "color": "0000ff"}})
			},
			want: []string{"bug", "enhancement", "question"},
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
					Reply(200).
					JSON([]map[string]string{})
			},
			want: []string{},
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
				t.Errorf("ListLabels() error = %v", err)
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
