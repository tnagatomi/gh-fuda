package executor

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tnagatomi/gh-fuda/internal/mock"
	"github.com/tnagatomi/gh-fuda/option"
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
		mock    *mock.MockAPI
		wantCall []struct {
			Label option.Label
			Repo  option.Repo
		}
		wantOut string
		wantErr bool
	}{
		{
			name:   "single repository",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement,question:0000ff:This is a question",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Created label "bug" for repository "tnagatomi/mock-repo"
Created label "enhancement" for repository "tnagatomi/mock-repo"
Created label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: option.Label{Name: "question", Description: "This is a question", Color: "0000ff"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
		},
		{
			name:   "multiple repository", 
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo-1,tnagatomi/mock-repo-2",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement,question:0000ff:This is a question",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Created label "bug" for repository "tnagatomi/mock-repo-1"
Created label "enhancement" for repository "tnagatomi/mock-repo-1"
Created label "question" for repository "tnagatomi/mock-repo-1"
Created label "bug" for repository "tnagatomi/mock-repo-2"
Created label "enhancement" for repository "tnagatomi/mock-repo-2"
Created label "question" for repository "tnagatomi/mock-repo-2"
`,
			wantErr: false,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: option.Label{Name: "question", Description: "This is a question", Color: "0000ff"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
				{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
				{Label: option.Label{Name: "question", Description: "This is a question", Color: "0000ff"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
			},
		},
		{
			name:   "dry-run",
			dryrun: true,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement,question:0000ff:This is a question",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Would create label "bug" for repository "tnagatomi/mock-repo"
Would create label "enhancement" for repository "tnagatomi/mock-repo"
Would create label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executor{
				api:    tt.mock,
				dryRun: tt.dryrun,
			}
			out := &bytes.Buffer{}
			err := e.Create(out, tt.args.repoOption, tt.args.labelOption)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Create() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if len(tt.wantCall) != len(tt.mock.CreateLabelCalls) {
				t.Errorf("Create() wantCall = %v, got %v", tt.wantCall, tt.mock.CreateLabelCalls)
			}
			for i, call := range tt.mock.CreateLabelCalls {
				if call.Label != tt.wantCall[i].Label || call.Repo != tt.wantCall[i].Repo {
					t.Errorf("Create() wantCall = %v, got %v", tt.wantCall, tt.mock.CreateLabelCalls)
				}
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
		mock    *mock.MockAPI
		wantOut string
		wantErr bool
		wantCall []struct {
			Label string
			Repo  option.Repo
		}
	}{
		{
			name:   "single repository",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug,enhancement,question",
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/mock-repo"
Deleted label "enhancement" for repository "tnagatomi/mock-repo"
Deleted label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
			wantCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: "question", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
		},
		{
			name:   "multiple repositories",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo-1,tnagatomi/mock-repo-2",
				labelOption: "bug,enhancement,question",
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/mock-repo-1"
Deleted label "enhancement" for repository "tnagatomi/mock-repo-1"
Deleted label "question" for repository "tnagatomi/mock-repo-1"
Deleted label "bug" for repository "tnagatomi/mock-repo-2"
Deleted label "enhancement" for repository "tnagatomi/mock-repo-2"
Deleted label "question" for repository "tnagatomi/mock-repo-2"
`,
			wantErr: false,
			wantCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: "question", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
				{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
				{Label: "question", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
			},
		},
		{
			name:   "dry-run",
			dryrun: true,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug,enhancement,question",
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Would delete label "bug" for repository "tnagatomi/mock-repo"
Would delete label "enhancement" for repository "tnagatomi/mock-repo"
Would delete label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
			wantCall: []struct {
				Label string
				Repo  option.Repo
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executor{
				api:    tt.mock,
				dryRun: tt.dryrun,
			}
			out := &bytes.Buffer{}
			err := e.Delete(out, tt.args.repoOption, tt.args.labelOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Delete() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if len(tt.wantCall) != len(tt.mock.DeleteLabelCalls) {
				t.Errorf("Delete() wantCall = %v, got %v", tt.wantCall, tt.mock.DeleteLabelCalls)
			}
			for i, call := range tt.mock.DeleteLabelCalls {
				if call.Label != tt.wantCall[i].Label || call.Repo != tt.wantCall[i].Repo {
					t.Errorf("Delete() wantCall = %v, got %v", tt.wantCall, tt.mock.DeleteLabelCalls)
				}
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
		mock    *mock.MockAPI
		wantOut string
		wantErr bool
		wantListCall []option.Repo
		wantCreateCall []struct {
			Label option.Label
			Repo  option.Repo
		}
		wantDeleteCall []struct {
			Label string
			Repo  option.Repo
		}
	}{
		{
			name:   "single repository",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					return []string{"bug", "enhancement", "question"}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
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
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo"},
			},
			wantCreateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: "question", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
		},
		{
			name:   "multiple repository",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo-1,tnagatomi/mock-repo-2",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					if repo.Owner == "tnagatomi" && repo.Repo == "mock-repo-1" {
						return []string{"bug", "enhancement", "question"}, nil
					} else if repo.Owner == "tnagatomi" && repo.Repo == "mock-repo-2" {
						return []string{"bug"}, nil
					}
					return nil, fmt.Errorf("unexpected repository: %v", repo)
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Emptying labels first
Deleted label "bug" for repository "tnagatomi/mock-repo-1"
Deleted label "enhancement" for repository "tnagatomi/mock-repo-1"
Deleted label "question" for repository "tnagatomi/mock-repo-1"
Deleted label "bug" for repository "tnagatomi/mock-repo-2"
Creating labels
Created label "bug" for repository "tnagatomi/mock-repo-1"
Created label "enhancement" for repository "tnagatomi/mock-repo-1"
Created label "bug" for repository "tnagatomi/mock-repo-2"
Created label "enhancement" for repository "tnagatomi/mock-repo-2"
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo-1"},
				{Owner: "tnagatomi", Repo: "mock-repo-2"},
			},
		wantCreateCall: []struct {
			Label option.Label
			Repo  option.Repo
		}{
			{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
			{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
			{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
			{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
		},
		wantDeleteCall: []struct {
			Label string
			Repo  option.Repo
		}{
			{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
			{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
			{Label: "question", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
			{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
		}},
		{
			name:   "dry-run",
			dryrun: true,
			args: args{
				repoOption:  "tnagatomi/mock-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					return []string{"bug", "enhancement", "question"}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Would delete label "bug" for repository "tnagatomi/mock-repo"
Would delete label "enhancement" for repository "tnagatomi/mock-repo"
Would delete label "question" for repository "tnagatomi/mock-repo"
Would create label "bug" for repository "tnagatomi/mock-repo"
Would create label "enhancement" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo"},
			},
			wantCreateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executor{
				api:    tt.mock,
				dryRun: tt.dryrun,
			}
			out := &bytes.Buffer{}
			err := e.Sync(out, tt.args.repoOption, tt.args.labelOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Sync() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if len(tt.wantListCall) != len(tt.mock.ListLabelsCalls) {
				t.Errorf("Sync() wantListCall = %v, got %v", tt.wantListCall, tt.mock.ListLabelsCalls)
			}
			for i, call := range tt.mock.ListLabelsCalls {
				if call.Repo != tt.wantListCall[i] {
					t.Errorf("Sync() wantListCall = %v, got %v", tt.wantListCall, tt.mock.ListLabelsCalls)
				}
			}
			if len(tt.wantCreateCall) != len(tt.mock.CreateLabelCalls) {
				t.Errorf("Sync() wantCreateCall = %v, got %v", tt.wantCreateCall, tt.mock.CreateLabelCalls)
			}
			for i, call := range tt.mock.CreateLabelCalls {
				if call.Label != tt.wantCreateCall[i].Label || call.Repo != tt.wantCreateCall[i].Repo {
					t.Errorf("Sync() wantCreateCall = %v, got %v", tt.wantCreateCall, tt.mock.CreateLabelCalls)
				}
			}
			if len(tt.wantDeleteCall) != len(tt.mock.DeleteLabelCalls) {
				t.Errorf("Sync() wantDeleteCall = %v, got %v", tt.wantDeleteCall, tt.mock.DeleteLabelCalls)
			}
			for i, call := range tt.mock.DeleteLabelCalls {
				if call.Label != tt.wantDeleteCall[i].Label || call.Repo != tt.wantDeleteCall[i].Repo {
					t.Errorf("Sync() wantDeleteCall = %v, got %v", tt.wantDeleteCall, tt.mock.DeleteLabelCalls)
				}
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
		mock    *mock.MockAPI
		wantOut string
		wantErr bool
		wantListCall []option.Repo
		wantDeleteCall []struct {
			Label string
			Repo  option.Repo
		}	
	}{
		{
			name:   "single repository",
			dryrun: false,
			args: args{
				repoOption: "tnagatomi/mock-repo",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					return []string{"bug", "enhancement", "question"}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/mock-repo"
Deleted label "enhancement" for repository "tnagatomi/mock-repo"
Deleted label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo"},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: "question", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
		},
		{
			name:   "multiple repository",
			dryrun: false,
			args: args{
				repoOption: "tnagatomi/mock-repo-1,tnagatomi/mock-repo-2",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					if repo.Owner == "tnagatomi" && repo.Repo == "mock-repo-1" {
						return []string{"bug", "enhancement", "question"}, nil
					} else if repo.Owner == "tnagatomi" && repo.Repo == "mock-repo-2" {
						return []string{"invalid", "help wanted"}, nil
					}
					return nil, fmt.Errorf("unexpected repository: %v", repo)
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/mock-repo-1"
Deleted label "enhancement" for repository "tnagatomi/mock-repo-1"
Deleted label "question" for repository "tnagatomi/mock-repo-1"
Deleted label "invalid" for repository "tnagatomi/mock-repo-2"
Deleted label "help wanted" for repository "tnagatomi/mock-repo-2"
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo-1"},
				{Owner: "tnagatomi", Repo: "mock-repo-2"},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: "question", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: "invalid", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
				{Label: "help wanted", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
			},
		},
		{
			name:   "dry-run",
			dryrun: true,
			args: args{
				repoOption: "tnagatomi/mock-repo",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					return []string{"bug", "enhancement", "question"}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Would delete label "bug" for repository "tnagatomi/mock-repo"
Would delete label "enhancement" for repository "tnagatomi/mock-repo"
Would delete label "question" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo"},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executor{
				api:    tt.mock,
				dryRun: tt.dryrun,
			}
			out := &bytes.Buffer{}
			err := e.Empty(out, tt.args.repoOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Sync() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if len(tt.wantListCall) != len(tt.mock.ListLabelsCalls) {
				t.Errorf("Empty() wantListCall = %v, got %v", tt.wantListCall, tt.mock.ListLabelsCalls)
			}
			for i, call := range tt.mock.ListLabelsCalls {
				if call.Repo != tt.wantListCall[i] {
					t.Errorf("Empty() wantListCall = %v, got %v", tt.wantListCall, tt.mock.ListLabelsCalls)
				}
			}
			if len(tt.wantDeleteCall) != len(tt.mock.DeleteLabelCalls) {
				t.Errorf("Empty() wantDeleteCall = %v, got %v", tt.wantDeleteCall, tt.mock.DeleteLabelCalls)
			}
			for i, call := range tt.mock.DeleteLabelCalls {
				if call.Label != tt.wantDeleteCall[i].Label || call.Repo != tt.wantDeleteCall[i].Repo {
					t.Errorf("Empty() wantDeleteCall = %v, got %v", tt.wantDeleteCall, tt.mock.DeleteLabelCalls)
				}
			}
		})
	}
}
