package executor

import (
	"bytes"
	"github.com/h2non/gock"
	"github.com/tnagatomi/gh-fuda/api"
	"net/http"
	"testing"
)

func TestCreate(t *testing.T) {
	type args struct {
		repoOption  string
		labelOption string
	}
	tests := []struct {
		name    string
		dryrun  bool
		args    args
		mock    func()
		wantOut string
		wantErr bool
	}{
		{
			name:   "create labels for single repository",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement,question:0000ff:This is a question",
			},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(201).
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"})
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}).
					Reply(201).
					JSON(map[string]string{"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"})
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "question", "description": "This is a question", "color": "0000ff"}).
					Reply(201).
					JSON(map[string]string{"name": "question", "description": "This is a question", "color": "0000ff"})
			},
			wantOut: `Created label "bug" for repository "tnagatomi/mock-repo"
Created label "enhancement" for repository "tnagatomi/mock-repo"
Created label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
		},
		{
			name:   "create labels for multiple repository",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo-1,tnagatomi/mock-repo-2",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement,question:0000ff:This is a question",
			},
			mock: func() {
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo-1/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(201).
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"})
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo-1/labels").
					MatchType("json").
					JSON(map[string]string{"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}).
					Reply(201).
					JSON(map[string]string{"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"})
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo-1/labels").
					MatchType("json").
					JSON(map[string]string{"name": "question", "description": "This is a question", "color": "0000ff"}).
					Reply(201).
					JSON(map[string]string{"name": "question", "description": "This is a question", "color": "0000ff"})
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo-2/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(201).
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"})
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo-2/labels").
					MatchType("json").
					JSON(map[string]string{"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}).
					Reply(201).
					JSON(map[string]string{"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"})
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo-2/labels").
					MatchType("json").
					JSON(map[string]string{"name": "question", "description": "This is a question", "color": "0000ff"}).
					Reply(201).
					JSON(map[string]string{"name": "question", "description": "This is a question", "color": "0000ff"})
			},
			wantOut: `Created label "bug" for repository "tnagatomi/mock-repo-1"
Created label "enhancement" for repository "tnagatomi/mock-repo-1"
Created label "question" for repository "tnagatomi/mock-repo-1"
Created label "bug" for repository "tnagatomi/mock-repo-2"
Created label "enhancement" for repository "tnagatomi/mock-repo-2"
Created label "question" for repository "tnagatomi/mock-repo-2"
`,
			wantErr: false,
		},
		{
			name:   "create labels dry-run",
			dryrun: true,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement,question:0000ff:This is a question",
			},
			mock: func() {},
			wantOut: `Would create label "bug" for repository "tnagatomi/mock-repo"
Would create label "enhancement" for repository "tnagatomi/mock-repo"
Would create label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			if tt.mock != nil {
				tt.mock()
			}

			api, err := api.NewAPI(http.DefaultClient)
			if err != nil {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			e := &Executor{
				api:    api,
				dryRun: tt.dryrun,
			}
			out := &bytes.Buffer{}
			err = e.Create(out, tt.args.repoOption, tt.args.labelOption)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Create() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		repoOption  string
		labelOption string
	}
	tests := []struct {
		name    string
		dryrun  bool
		args    args
		mock    func()
		wantOut string
		wantErr bool
	}{
		{
			name:   "delete labels",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug,enhancement,question",
			},
			mock: func() {
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/bug").
					Reply(204)
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/enhancement").
					Reply(204)
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/question").
					Reply(204)
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/mock-repo"
Deleted label "enhancement" for repository "tnagatomi/mock-repo"
Deleted label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
		},
		{
			name:   "delete labels dry-run",
			dryrun: true,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug,enhancement,question",
			},
			mock: func() {},
			wantOut: `Would delete label "bug" for repository "tnagatomi/mock-repo"
Would delete label "enhancement" for repository "tnagatomi/mock-repo"
Would delete label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			if tt.mock != nil {
				tt.mock()
			}

			api, err := api.NewAPI(http.DefaultClient)
			if err != nil {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			e := &Executor{
				api:    api,
				dryRun: tt.dryrun,
			}
			out := &bytes.Buffer{}
			err = e.Delete(out, tt.args.repoOption, tt.args.labelOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Delete() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestSync(t *testing.T) {
	type args struct {
		repoOption  string
		labelOption string
	}
	tests := []struct {
		name    string
		dryrun  bool
		args    args
		mock    func()
		wantOut string
		wantErr bool
	}{
		{
			name:   "sync labels",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: func() {
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/mock-repo/labels").
					Reply(200).
					JSON([]map[string]string{{"name": "bug", "description": "This is a bug", "color": "ff0000"}, {"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}, {"name": "question", "description": "This is a question", "color": "0000ff"}})
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/bug").
					Reply(204)
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/enhancement").
					Reply(204)
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/question").
					Reply(204)
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"}).
					Reply(201).
					JSON(map[string]string{"name": "bug", "description": "This is a bug", "color": "ff0000"})
				gock.New("https://api.github.com").
					Post("/repos/tnagatomi/mock-repo/labels").
					MatchType("json").
					JSON(map[string]string{"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}).
					Reply(201).
					JSON(map[string]string{"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"})
			},
			wantOut: `Emptying labels first
Deleted label "bug" for repository "tnagatomi/mock-repo"
Deleted label "enhancement" for repository "tnagatomi/mock-repo"
Deleted label "question" for repository "tnagatomi/mock-repo"
Creating labels
Created label "bug" for repository "tnagatomi/mock-repo"
Created label "enhancement" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
		},
		{
			name:   "sync labels dry-run",
			dryrun: true,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: func() {
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/mock-repo/labels").
					Reply(200).
					JSON([]map[string]string{{"name": "bug", "description": "This is a bug", "color": "ff0000"}, {"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}, {"name": "question", "description": "This is a question", "color": "0000ff"}})
			},
			wantOut: `Would delete label "bug" for repository "tnagatomi/mock-repo"
Would delete label "enhancement" for repository "tnagatomi/mock-repo"
Would delete label "question" for repository "tnagatomi/mock-repo"
Would create label "bug" for repository "tnagatomi/mock-repo"
Would create label "enhancement" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			if tt.mock != nil {
				tt.mock()
			}

			api, err := api.NewAPI(http.DefaultClient)
			if err != nil {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			e := &Executor{
				api:    api,
				dryRun: tt.dryrun,
			}
			out := &bytes.Buffer{}
			err = e.Sync(out, tt.args.repoOption, tt.args.labelOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Sync() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}

func TestEmpty(t *testing.T) {
	type args struct {
		repoOption string
	}
	tests := []struct {
		name    string
		dryrun  bool
		args    args
		mock    func()
		wantOut string
		wantErr bool
	}{
		{
			name:   "empty labels",
			dryrun: false,
			args: args{
				repoOption: "tnagatomi/mock-repo",
			},
			mock: func() {
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/mock-repo/labels").
					Reply(200).
					JSON([]map[string]string{{"name": "bug", "description": "This is a bug", "color": "ff0000"}, {"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}, {"name": "question", "description": "This is a question", "color": "0000ff"}})
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/bug").
					Reply(204)
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/enhancement").
					Reply(204)
				gock.New("https://api.github.com").
					Delete("/repos/tnagatomi/mock-repo/labels/question").
					Reply(204)
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/mock-repo"
Deleted label "enhancement" for repository "tnagatomi/mock-repo"
Deleted label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
		},
		{
			name:   "empty labels dry-run",
			dryrun: true,
			args: args{
				repoOption: "tnagatomi/mock-repo",
			},
			mock: func() {
				gock.New("https://api.github.com").
					Get("/repos/tnagatomi/mock-repo/labels").
					Reply(200).
					JSON([]map[string]string{{"name": "bug", "description": "This is a bug", "color": "ff0000"}, {"name": "enhancement", "description": "This is an enhancement", "color": "00ff00"}, {"name": "question", "description": "This is a question", "color": "0000ff"}})
			},
			wantOut: `Would delete label "bug" for repository "tnagatomi/mock-repo"
Would delete label "enhancement" for repository "tnagatomi/mock-repo"
Would delete label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			if tt.mock != nil {
				tt.mock()
			}

			api, err := api.NewAPI(http.DefaultClient)
			if err != nil {
				t.Errorf("Empty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			e := &Executor{
				api:    api,
				dryRun: tt.dryrun,
			}
			out := &bytes.Buffer{}
			err = e.Empty(out, tt.args.repoOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Sync() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if !gock.IsDone() {
				t.Errorf("pending mocks: %d", len(gock.Pending()))
			}
		})
	}
}
