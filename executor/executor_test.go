package executor

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tnagatomi/gh-fuda/api"
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

Summary: all operations completed successfully
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

Summary: all operations completed successfully
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
			name:   "repository not found error",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/non-existent-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.NotFoundError{Resource: fmt.Sprintf("repository %q", repo.String())}
				},
			},
			wantOut: `Failed to create label "bug" for repository "tnagatomi/non-existent-repo": repository "tnagatomi/non-existent-repo" not found
Failed to create label "enhancement" for repository "tnagatomi/non-existent-repo": repository "tnagatomi/non-existent-repo" not found

Summary: some operations failed: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "non-existent-repo"}},
				{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "non-existent-repo"}},
			},
		},
		{
			name:   "permission error",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/private-repo",
				labelOption: "bug:ff0000:This is a bug",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to create label "bug" for repository "tnagatomi/private-repo": forbidden

Summary: some operations failed: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "private-repo"}},
			},
		},
		{
			name:   "multiple repositories with partial failure",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/repo-1,tnagatomi/repo-2,tnagatomi/repo-3",
				labelOption: "bug:ff0000:This is a bug",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					if repo.Repo == "repo-2" {
						return &api.NotFoundError{Resource: fmt.Sprintf("repository %q", repo.String())}
					}
					return nil
				},
			},
			wantOut: `Created label "bug" for repository "tnagatomi/repo-1"
Failed to create label "bug" for repository "tnagatomi/repo-2": repository "tnagatomi/repo-2" not found
Created label "bug" for repository "tnagatomi/repo-3"

Summary: some operations failed: 2 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-1"}},
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-2"}},
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-3"}},
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

Summary: all operations completed successfully
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

Summary: all operations completed successfully
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
			name:   "repository not found error",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/non-existent-repo",
				labelOption: "bug,enhancement",
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return &api.NotFoundError{Resource: fmt.Sprintf("repository %q", repo.String())}
				},
			},
			wantOut: `Failed to delete label "bug" for repository "tnagatomi/non-existent-repo": repository "tnagatomi/non-existent-repo" not found
Failed to delete label "enhancement" for repository "tnagatomi/non-existent-repo": repository "tnagatomi/non-existent-repo" not found

Summary: some operations failed: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "non-existent-repo"}},
				{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "non-existent-repo"}},
			},
		},
		{
			name:   "permission error",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/private-repo",
				labelOption: "bug",
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to delete label "bug" for repository "tnagatomi/private-repo": forbidden

Summary: some operations failed: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "private-repo"}},
			},
		},
		{
			name:   "multiple repositories with partial failure",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/repo-1,tnagatomi/repo-2,tnagatomi/repo-3",
				labelOption: "bug",
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					if repo.Repo == "repo-2" {
						return &api.NotFoundError{Resource: fmt.Sprintf("repository %q", repo.String())}
					}
					return nil
				},
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/repo-1"
Failed to delete label "bug" for repository "tnagatomi/repo-2": repository "tnagatomi/repo-2" not found
Deleted label "bug" for repository "tnagatomi/repo-3"

Summary: some operations failed: 2 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-1"}},
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-2"}},
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-3"}},
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
		wantUpdateCall []struct {
			Label option.Label
			Repo  option.Repo
		}
		wantDeleteCall []struct {
			Label string
			Repo  option.Repo
		}
	}{
		{
			name:   "multiple repositories with different existing labels",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo-1,tnagatomi/mock-repo-2",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					if repo.Repo == "mock-repo-1" {
						return []string{"bug", "enhancement", "question"}, nil
					}
					return []string{"bug", "help wanted"}, nil
				},
				UpdateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Deleted label "question" for repository "tnagatomi/mock-repo-1"
Updated label "bug" for repository "tnagatomi/mock-repo-1"
Updated label "enhancement" for repository "tnagatomi/mock-repo-1"
Deleted label "help wanted" for repository "tnagatomi/mock-repo-2"
Updated label "bug" for repository "tnagatomi/mock-repo-2"
Created label "enhancement" for repository "tnagatomi/mock-repo-2"

Summary: all operations completed successfully
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
				{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
			},
			wantUpdateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "question", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: "help wanted", Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-2"}},
			},
		},
		{
			name:   "repository not found on list labels",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/mock-repo-1,tnagatomi/non-existent-repo",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					if repo.Repo == "non-existent-repo" {
						return nil, &api.NotFoundError{Resource: fmt.Sprintf("repository %q", repo.String())}
					}
					return []string{}, nil
				},
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Created label "bug" for repository "tnagatomi/mock-repo-1"
Created label "enhancement" for repository "tnagatomi/mock-repo-1"
Failed to list labels for repository "tnagatomi/non-existent-repo": repository "tnagatomi/non-existent-repo" not found

Summary: some operations failed: 1 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo-1"},
				{Owner: "tnagatomi", Repo: "non-existent-repo"},
			},
			wantCreateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
				{Label: option.Label{Name: "enhancement", Description: "This is an enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo-1"}},
			},
			wantUpdateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{},
		},
		{
			name:   "permission error on update",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/private-repo-1,tnagatomi/private-repo-2",
				labelOption: "bug:ff0000:This is a bug",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					return []string{"bug"}, nil
				},
				UpdateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to update label "bug" for repository "tnagatomi/private-repo-1": forbidden
Failed to update label "bug" for repository "tnagatomi/private-repo-2": forbidden

Summary: some operations failed: 0 repositories succeeded, 2 failed
`,
			wantErr: true,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "private-repo-1"},
				{Owner: "tnagatomi", Repo: "private-repo-2"},
			},
			wantCreateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{},
			wantUpdateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "private-repo-1"}},
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "private-repo-2"}},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{},
		},
		{
			name:   "mixed errors during sync",
			dryrun: false,
			args: args{
				repoOption:  "tnagatomi/repo-1,tnagatomi/repo-2,tnagatomi/repo-3",
				labelOption: "bug:ff0000:This is a bug",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					if repo.Repo == "repo-2" {
						return nil, &api.NotFoundError{Resource: fmt.Sprintf("repository %q", repo.String())}
					}
					return []string{"old-label"}, nil
				},
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					if repo.Repo == "repo-3" {
						return &api.ForbiddenError{}
					}
					return nil
				},
			},
			wantOut: `Deleted label "old-label" for repository "tnagatomi/repo-1"
Created label "bug" for repository "tnagatomi/repo-1"
Failed to list labels for repository "tnagatomi/repo-2": repository "tnagatomi/repo-2" not found
Failed to delete label "old-label" for repository "tnagatomi/repo-3": forbidden
Created label "bug" for repository "tnagatomi/repo-3"

Summary: some operations failed: 1 repositories succeeded, 2 failed
`,
			wantErr: true,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "repo-1"},
				{Owner: "tnagatomi", Repo: "repo-2"},
				{Owner: "tnagatomi", Repo: "repo-3"},
			},
			wantCreateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-1"}},
				{Label: option.Label{Name: "bug", Description: "This is a bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-3"}},
			},
			wantUpdateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "old-label", Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-1"}},
				{Label: "old-label", Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-3"}},
			},
		},
		{
			name:   "dry-run",
			dryrun: true,
			args: args{
				repoOption:  "tnagatomi/mock-repo-1,tnagatomi/mock-repo-2",
				labelOption: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					if repo.Repo == "mock-repo-1" {
						return []string{"bug", "enhancement", "question"}, nil
					}
					return []string{}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Would delete label "question" for repository "tnagatomi/mock-repo-1"
Would update label "bug" for repository "tnagatomi/mock-repo-1"
Would update label "enhancement" for repository "tnagatomi/mock-repo-1"
Would create label "bug" for repository "tnagatomi/mock-repo-2"
Would create label "enhancement" for repository "tnagatomi/mock-repo-2"
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo-1"},
				{Owner: "tnagatomi", Repo: "mock-repo-2"},
			},
			wantCreateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{},
			wantUpdateCall: []struct {
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
			if len(tt.wantUpdateCall) != len(tt.mock.UpdateLabelCalls) {
				t.Errorf("Sync() wantUpdateCall = %v, got %v", tt.wantUpdateCall, tt.mock.UpdateLabelCalls)
			}
			for i, call := range tt.mock.UpdateLabelCalls {
				if call.Label != tt.wantUpdateCall[i].Label || call.Repo != tt.wantUpdateCall[i].Repo {
					t.Errorf("Sync() wantUpdateCall = %v, got %v", tt.wantUpdateCall, tt.mock.UpdateLabelCalls)
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

Summary: all operations completed successfully
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

Summary: all operations completed successfully
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
			name:   "repository not found on list",
			dryrun: false,
			args: args{
				repoOption: "tnagatomi/non-existent-repo",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					return nil, &api.NotFoundError{Resource: fmt.Sprintf("repository %q", repo.String())}
				},
			},
			wantOut: `Failed to list labels for repository "tnagatomi/non-existent-repo": repository "tnagatomi/non-existent-repo" not found

Summary: some operations failed: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "non-existent-repo"},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{},
		},
		{
			name:   "permission error on delete",
			dryrun: false,
			args: args{
				repoOption: "tnagatomi/private-repo",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					return []string{"bug", "enhancement"}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to delete label "bug" for repository "tnagatomi/private-repo": forbidden
Failed to delete label "enhancement" for repository "tnagatomi/private-repo": forbidden

Summary: some operations failed: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "private-repo"},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "private-repo"}},
				{Label: "enhancement", Repo: option.Repo{Owner: "tnagatomi", Repo: "private-repo"}},
			},
		},
		{
			name:   "multiple repositories with partial failure",
			dryrun: false,
			args: args{
				repoOption: "tnagatomi/repo-1,tnagatomi/repo-2,tnagatomi/repo-3",
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]string, error) {
					if repo.Repo == "repo-2" {
						return nil, &api.NotFoundError{Resource: fmt.Sprintf("repository %q", repo.String())}
					}
					return []string{"bug"}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/repo-1"
Failed to list labels for repository "tnagatomi/repo-2": repository "tnagatomi/repo-2" not found
Deleted label "bug" for repository "tnagatomi/repo-3"

Summary: some operations failed: 2 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "repo-1"},
				{Owner: "tnagatomi", Repo: "repo-2"},
				{Owner: "tnagatomi", Repo: "repo-3"},
			},
			wantDeleteCall: []struct {
				Label string
				Repo  option.Repo
			}{
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-1"}},
				{Label: "bug", Repo: option.Repo{Owner: "tnagatomi", Repo: "repo-3"}},
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
