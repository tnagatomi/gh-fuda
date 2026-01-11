package executor

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/tnagatomi/gh-fuda/api"
	"github.com/tnagatomi/gh-fuda/internal/mock"
	"github.com/tnagatomi/gh-fuda/option"
)

// stripProgress removes progress output (e.g., "Progress: X/Y completed") and ANSI escape codes from output
func stripProgress(s string) string {
	// Remove ANSI escape codes (like \033[K for clear line)
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	s = ansiRegex.ReplaceAllString(s, "")
	// Remove progress lines (Progress: X/Y completed with optional \r)
	progressRegex := regexp.MustCompile(`\r?Progress: \d+/\d+ completed`)
	s = progressRegex.ReplaceAllString(s, "")
	// Remove any remaining carriage returns
	s = regexp.MustCompile(`\r`).ReplaceAllString(s, "")
	return s
}

// sortedLines returns a sorted list of non-empty, non-summary lines from output
func sortedLines(s string) []string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		// Skip empty lines and summary lines
		if line == "" || strings.HasPrefix(line, "Summary:") {
			continue
		}
		result = append(result, line)
	}
	sort.Strings(result)
	return result
}

// containsAllCalls checks if got contains all expected calls (order independent)
func containsAllCalls(got, want []struct {
	Label option.Label
	Repo  option.Repo
}) bool {
	if len(got) != len(want) {
		return false
	}
	// Create a map to track which expected calls have been found
	found := make([]bool, len(want))
	for _, g := range got {
		for i, w := range want {
			if !found[i] && g.Label == w.Label && g.Repo == w.Repo {
				found[i] = true
				break
			}
		}
	}
	for _, f := range found {
		if !f {
			return false
		}
	}
	return true
}

// containsAllDeleteCalls checks if got contains all expected delete calls (order independent)
func containsAllDeleteCalls(got, want []struct {
	Label string
	Repo  option.Repo
}) bool {
	if len(got) != len(want) {
		return false
	}
	found := make([]bool, len(want))
	for _, g := range got {
		for i, w := range want {
			if !found[i] && g.Label == w.Label && g.Repo == w.Repo {
				found[i] = true
				break
			}
		}
	}
	for _, f := range found {
		if !f {
			return false
		}
	}
	return true
}

// containsAllListCalls checks if got contains all expected list calls (order independent)
func containsAllListCalls(got []struct{ Repo option.Repo }, want []option.Repo) bool {
	if len(got) != len(want) {
		return false
	}
	found := make([]bool, len(want))
	for _, g := range got {
		for i, w := range want {
			if !found[i] && g.Repo == w {
				found[i] = true
				break
			}
		}
	}
	for _, f := range found {
		if !f {
			return false
		}
	}
	return true
}

func TestCreate(t *testing.T) {
	type args struct {
		repos  []option.Repo
		labels []option.Label
	}
	tests := []struct {
		name    string
		dryrun  bool
		force   bool
		args    args
		mock    *mock.MockAPI
		wantCall []struct {
			Label option.Label
			Repo  option.Repo
		}
		wantUpdateCall []struct {
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
					{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
					{Name: "question", Color: "0000ff", Description: "This is a question"},
				},
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo-1"},
					{Owner: "tnagatomi", Repo: "mock-repo-2"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
					{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
					{Name: "question", Color: "0000ff", Description: "This is a question"},
				},
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "non-existent-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
					{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
				},
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
				},
			},
			wantOut: `Failed to create label "bug" for repository "tnagatomi/non-existent-repo": repository not found
Failed to create label "enhancement" for repository "tnagatomi/non-existent-repo": repository not found

Summary: 0 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "private-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
				},
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to create label "bug" for repository "tnagatomi/private-repo": forbidden

Summary: 0 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "repo-1"},
					{Owner: "tnagatomi", Repo: "repo-2"},
					{Owner: "tnagatomi", Repo: "repo-3"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
				},
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					if repo.Repo == "repo-2" {
						return &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
					}
					return nil
				},
			},
			wantOut: `Created label "bug" for repository "tnagatomi/repo-1"
Failed to create label "bug" for repository "tnagatomi/repo-2": repository not found
Created label "bug" for repository "tnagatomi/repo-3"

Summary: 2 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
					{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
					{Name: "question", Color: "0000ff", Description: "This is a question"},
				},
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
		{
			name:   "force flag - update existing label",
			dryrun: false,
			force:  true,
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "Updated description"},
				},
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.AlreadyExistsError{ResourceType: api.ResourceTypeLabel}
				},
				UpdateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Updated label "bug" for repository "tnagatomi/mock-repo"

Summary: all operations completed successfully
`,
			wantErr: false,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "Updated description", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
			wantUpdateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "Updated description", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
		},
		{
			name:   "force flag false - fail on existing label",
			dryrun: false,
			force:  false,
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "Description"},
				},
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.AlreadyExistsError{ResourceType: api.ResourceTypeLabel}
				},
			},
			wantOut: `Failed to create label "bug" for repository "tnagatomi/mock-repo": label already exists

Summary: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "Description", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
		},
		{
			name:   "force flag - update fails",
			dryrun: false,
			force:  true,
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "Description"},
				},
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.AlreadyExistsError{ResourceType: api.ResourceTypeLabel}
				},
				UpdateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to update label "bug" for repository "tnagatomi/mock-repo": forbidden

Summary: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "Description", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
			wantUpdateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "Description", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
		},
		{
			name:   "force flag - mixed create and update",
			dryrun: false,
			force:  true,
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "Bug"},
					{Name: "enhancement", Color: "00ff00", Description: "Enhancement"},
				},
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					if label.Name == "bug" {
						return &api.AlreadyExistsError{ResourceType: api.ResourceTypeLabel}
					}
					return nil
				},
				UpdateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Updated label "bug" for repository "tnagatomi/mock-repo"
Created label "enhancement" for repository "tnagatomi/mock-repo"

Summary: all operations completed successfully
`,
			wantErr: false,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "Bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
				{Label: option.Label{Name: "enhancement", Description: "Enhancement", Color: "00ff00"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
			wantUpdateCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{
				{Label: option.Label{Name: "bug", Description: "Bug", Color: "ff0000"}, Repo: option.Repo{Owner: "tnagatomi", Repo: "mock-repo"}},
			},
		},
		{
			name:   "dry-run with force flag - existing labels show update message",
			dryrun: true,
			force:  true,
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "Bug"},
					{Name: "enhancement", Color: "00ff00", Description: "Enhancement"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return []option.Label{{Name: "bug"}}, nil
				},
			},
			wantOut: `Would update label "bug" for repository "tnagatomi/mock-repo"
Would create label "enhancement" for repository "tnagatomi/mock-repo"
`,
			wantErr: false,
			wantCall: []struct {
				Label option.Label
				Repo  option.Repo
			}{},
		},
		{
			name:   "dry-run with force flag - list labels fails",
			dryrun: true,
			force:  true,
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "non-existent-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "Bug"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return nil, &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
				},
			},
			wantOut: `Failed to list labels for repository "tnagatomi/non-existent-repo": repository not found
`,
			wantErr: true,
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
			err := e.Create(out, tt.args.repos, tt.args.labels, tt.force)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotOut := stripProgress(out.String())
			// For parallel execution, compare output lines in sorted order
			if sortedLines(gotOut) != nil && sortedLines(tt.wantOut) != nil {
				gotLines := sortedLines(gotOut)
				wantLines := sortedLines(tt.wantOut)
				if len(gotLines) != len(wantLines) {
					t.Errorf("Create() gotOut = %q, want %q", gotOut, tt.wantOut)
				} else {
					for i := range gotLines {
						if gotLines[i] != wantLines[i] {
							t.Errorf("Create() gotOut = %q, want %q", gotOut, tt.wantOut)
							break
						}
					}
				}
			}
			// Use order-independent comparison for API calls (parallel execution)
			if !containsAllCalls(tt.mock.CreateLabelCalls, tt.wantCall) {
				t.Errorf("Create() wantCall = %v, got %v", tt.wantCall, tt.mock.CreateLabelCalls)
			}
			if !containsAllCalls(tt.mock.UpdateLabelCalls, tt.wantUpdateCall) {
				t.Errorf("Create() wantUpdateCall = %v, got %v", tt.wantUpdateCall, tt.mock.UpdateLabelCalls)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		repos  []option.Repo
		labels []string
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []string{"bug", "enhancement", "question"},
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo-1"},
					{Owner: "tnagatomi", Repo: "mock-repo-2"},
				},
				labels: []string{"bug", "enhancement", "question"},
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "non-existent-repo"},
				},
				labels: []string{"bug", "enhancement"},
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
				},
			},
			wantOut: `Failed to delete label "bug" for repository "tnagatomi/non-existent-repo": repository not found
Failed to delete label "enhancement" for repository "tnagatomi/non-existent-repo": repository not found

Summary: 0 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "private-repo"},
				},
				labels: []string{"bug"},
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to delete label "bug" for repository "tnagatomi/private-repo": forbidden

Summary: 0 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "repo-1"},
					{Owner: "tnagatomi", Repo: "repo-2"},
					{Owner: "tnagatomi", Repo: "repo-3"},
				},
				labels: []string{"bug"},
			},
			mock: &mock.MockAPI{
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					if repo.Repo == "repo-2" {
						return &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
					}
					return nil
				},
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/repo-1"
Failed to delete label "bug" for repository "tnagatomi/repo-2": repository not found
Deleted label "bug" for repository "tnagatomi/repo-3"

Summary: 2 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
				labels: []string{"bug", "enhancement", "question"},
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
			err := e.Delete(out, tt.args.repos, tt.args.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotOut := stripProgress(out.String())
			// For parallel execution, compare output lines in sorted order
			if sortedLines(gotOut) != nil && sortedLines(tt.wantOut) != nil {
				gotLines := sortedLines(gotOut)
				wantLines := sortedLines(tt.wantOut)
				if len(gotLines) != len(wantLines) {
					t.Errorf("Delete() gotOut = %q, want %q", gotOut, tt.wantOut)
				} else {
					for i := range gotLines {
						if gotLines[i] != wantLines[i] {
							t.Errorf("Delete() gotOut = %q, want %q", gotOut, tt.wantOut)
							break
						}
					}
				}
			}
			// Use order-independent comparison for API calls (parallel execution)
			if !containsAllDeleteCalls(tt.mock.DeleteLabelCalls, tt.wantCall) {
				t.Errorf("Delete() wantCall = %v, got %v", tt.wantCall, tt.mock.DeleteLabelCalls)
			}
		})
	}
}

func TestSync(t *testing.T) {
	type args struct {
		repos  []option.Repo
		labels []option.Label
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo-1"},
					{Owner: "tnagatomi", Repo: "mock-repo-2"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
					{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					if repo.Repo == "mock-repo-1" {
						return []option.Label{
						{Name: "bug"},
						{Name: "enhancement"},
						{Name: "question"},
					}, nil
					}
					return []option.Label{
						{Name: "bug"},
						{Name: "help wanted"},
					}, nil
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo-1"},
					{Owner: "tnagatomi", Repo: "non-existent-repo"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
					{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					if repo.Repo == "non-existent-repo" {
						return nil, &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
					}
					return []option.Label{}, nil
				},
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Created label "bug" for repository "tnagatomi/mock-repo-1"
Created label "enhancement" for repository "tnagatomi/mock-repo-1"
Failed to list labels for repository "tnagatomi/non-existent-repo": repository not found

Summary: 1 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "private-repo-1"},
					{Owner: "tnagatomi", Repo: "private-repo-2"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return []option.Label{{Name: "bug"}}, nil
				},
				UpdateLabelFunc: func(label option.Label, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to update label "bug" for repository "tnagatomi/private-repo-1": forbidden
Failed to update label "bug" for repository "tnagatomi/private-repo-2": forbidden

Summary: 0 repositories succeeded, 2 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "repo-1"},
					{Owner: "tnagatomi", Repo: "repo-2"},
					{Owner: "tnagatomi", Repo: "repo-3"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					if repo.Repo == "repo-2" {
						return nil, &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
					}
					return []option.Label{{Name: "old-label"}}, nil
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
Failed to list labels for repository "tnagatomi/repo-2": repository not found
Failed to delete label "old-label" for repository "tnagatomi/repo-3": forbidden
Created label "bug" for repository "tnagatomi/repo-3"

Summary: 1 repositories succeeded, 2 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo-1"},
					{Owner: "tnagatomi", Repo: "mock-repo-2"},
				},
				labels: []option.Label{
					{Name: "bug", Color: "ff0000", Description: "This is a bug"},
					{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
				},
			},
			mock: &mock.MockAPI{
				CreateLabelFunc: func(label option.Label, repo option.Repo) error {
					return nil
				},
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					if repo.Repo == "mock-repo-1" {
						return []option.Label{
						{Name: "bug"},
						{Name: "enhancement"},
						{Name: "question"},
					}, nil
					}
					return []option.Label{}, nil
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
			err := e.Sync(out, tt.args.repos, tt.args.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotOut := stripProgress(out.String())
			// For parallel execution, compare output lines in sorted order
			if sortedLines(gotOut) != nil && sortedLines(tt.wantOut) != nil {
				gotLines := sortedLines(gotOut)
				wantLines := sortedLines(tt.wantOut)
				if len(gotLines) != len(wantLines) {
					t.Errorf("Sync() gotOut = %q, want %q", gotOut, tt.wantOut)
				} else {
					for i := range gotLines {
						if gotLines[i] != wantLines[i] {
							t.Errorf("Sync() gotOut = %q, want %q", gotOut, tt.wantOut)
							break
						}
					}
				}
			}
			// Use order-independent comparison for API calls (parallel execution)
			if !containsAllListCalls(tt.mock.ListLabelsCalls, tt.wantListCall) {
				t.Errorf("Sync() wantListCall = %v, got %v", tt.wantListCall, tt.mock.ListLabelsCalls)
			}
			if !containsAllCalls(tt.mock.CreateLabelCalls, tt.wantCreateCall) {
				t.Errorf("Sync() wantCreateCall = %v, got %v", tt.wantCreateCall, tt.mock.CreateLabelCalls)
			}
			if !containsAllDeleteCalls(tt.mock.DeleteLabelCalls, tt.wantDeleteCall) {
				t.Errorf("Sync() wantDeleteCall = %v, got %v", tt.wantDeleteCall, tt.mock.DeleteLabelCalls)
			}
			if !containsAllCalls(tt.mock.UpdateLabelCalls, tt.wantUpdateCall) {
				t.Errorf("Sync() wantUpdateCall = %v, got %v", tt.wantUpdateCall, tt.mock.UpdateLabelCalls)
			}
		})
	}
}

func TestList(t *testing.T) {
	type args struct {
		repos []option.Repo
	}
	tests := []struct {
		name    string
		args    args
		mock    *mock.MockAPI
		wantOut string
		wantErr bool
		wantListCall []option.Repo
	}{
		{
			name: "single repository with labels",
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return []option.Label{
						{Name: "bug", Color: "d73a4a", Description: "Something isn't working"},
						{Name: "enhancement", Color: "a2eeef", Description: "New feature or request"},
						{Name: "documentation", Color: "0075ca", Description: ""},
					}, nil
				},
			},
			wantOut: `Labels for repository "tnagatomi/mock-repo":
  bug (#d73a4a) - Something isn't working
  enhancement (#a2eeef) - New feature or request
  documentation (#0075ca)

Summary: all operations completed successfully
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo"},
			},
		},
		{
			name: "multiple repositories",
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo-1"},
					{Owner: "tnagatomi", Repo: "mock-repo-2"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					if repo.Repo == "mock-repo-1" {
						return []option.Label{
							{Name: "bug", Color: "d73a4a", Description: "Something isn't working"},
						}, nil
					}
					return []option.Label{
						{Name: "enhancement", Color: "a2eeef", Description: ""},
						{Name: "good first issue", Color: "7057ff", Description: "Good for newcomers"},
					}, nil
				},
			},
			wantOut: `Labels for repository "tnagatomi/mock-repo-1":
  bug (#d73a4a) - Something isn't working
Labels for repository "tnagatomi/mock-repo-2":
  enhancement (#a2eeef)
  good first issue (#7057ff) - Good for newcomers

Summary: all operations completed successfully
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "mock-repo-1"},
				{Owner: "tnagatomi", Repo: "mock-repo-2"},
			},
		},
		{
			name: "repository with no labels",
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "empty-repo"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return []option.Label{}, nil
				},
			},
			wantOut: `Repository "tnagatomi/empty-repo" has no labels

Summary: all operations completed successfully
`,
			wantErr: false,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "empty-repo"},
			},
		},
		{
			name: "repository not found",
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "non-existent-repo"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return nil, &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
				},
			},
			wantOut: `Failed to list labels for repository "tnagatomi/non-existent-repo": repository not found

Summary: 0 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "non-existent-repo"},
			},
		},
		{
			name: "mixed success and failure",
			args: args{
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "repo-1"},
					{Owner: "tnagatomi", Repo: "repo-2"},
					{Owner: "tnagatomi", Repo: "repo-3"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					if repo.Repo == "repo-2" {
						return nil, &api.ForbiddenError{}
					}
					return []option.Label{
						{Name: "bug", Color: "d73a4a", Description: ""},
					}, nil
				},
			},
			wantOut: `Labels for repository "tnagatomi/repo-1":
  bug (#d73a4a)
Failed to list labels for repository "tnagatomi/repo-2": forbidden
Labels for repository "tnagatomi/repo-3":
  bug (#d73a4a)

Summary: 2 repositories succeeded, 1 failed
`,
			wantErr: true,
			wantListCall: []option.Repo{
				{Owner: "tnagatomi", Repo: "repo-1"},
				{Owner: "tnagatomi", Repo: "repo-2"},
				{Owner: "tnagatomi", Repo: "repo-3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executor{
				api:    tt.mock,
				dryRun: false,
			}
			out := &bytes.Buffer{}
			err := e.List(out, tt.args.repos)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := stripProgress(out.String()); gotOut != tt.wantOut {
				t.Errorf("List() gotOut = %q, want %q", gotOut, tt.wantOut)
			}
			// Use order-independent comparison for API calls (parallel execution)
			if !containsAllListCalls(tt.mock.ListLabelsCalls, tt.wantListCall) {
				t.Errorf("List() wantListCall = %v, got %v", tt.wantListCall, tt.mock.ListLabelsCalls)
			}
		})
	}
}

func TestEmpty(t *testing.T) {
	type args struct {
		repos []option.Repo
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return []option.Label{
						{Name: "bug"},
						{Name: "enhancement"},
						{Name: "question"},
					}, nil
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo-1"},
					{Owner: "tnagatomi", Repo: "mock-repo-2"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					if repo.Owner == "tnagatomi" && repo.Repo == "mock-repo-1" {
						return []option.Label{
						{Name: "bug"},
						{Name: "enhancement"},
						{Name: "question"},
					}, nil
					} else if repo.Owner == "tnagatomi" && repo.Repo == "mock-repo-2" {
						return []option.Label{
						{Name: "invalid"},
						{Name: "help wanted"},
					}, nil
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "non-existent-repo"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return nil, &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
				},
			},
			wantOut: `Failed to list labels for repository "tnagatomi/non-existent-repo": repository not found

Summary: 0 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "private-repo"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return []option.Label{
					{Name: "bug"},
					{Name: "enhancement"},
				}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return &api.ForbiddenError{}
				},
			},
			wantOut: `Failed to delete label "bug" for repository "tnagatomi/private-repo": forbidden
Failed to delete label "enhancement" for repository "tnagatomi/private-repo": forbidden

Summary: 0 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "repo-1"},
					{Owner: "tnagatomi", Repo: "repo-2"},
					{Owner: "tnagatomi", Repo: "repo-3"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					if repo.Repo == "repo-2" {
						return nil, &api.NotFoundError{ResourceType: api.ResourceTypeRepository}
					}
					return []option.Label{{Name: "bug"}}, nil
				},
				DeleteLabelFunc: func(label string, repo option.Repo) error {
					return nil
				},
			},
			wantOut: `Deleted label "bug" for repository "tnagatomi/repo-1"
Failed to list labels for repository "tnagatomi/repo-2": repository not found
Deleted label "bug" for repository "tnagatomi/repo-3"

Summary: 2 repositories succeeded, 1 failed
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
				repos: []option.Repo{
					{Owner: "tnagatomi", Repo: "mock-repo"},
				},
			},
			mock: &mock.MockAPI{
				ListLabelsFunc: func(repo option.Repo) ([]option.Label, error) {
					return []option.Label{
						{Name: "bug"},
						{Name: "enhancement"},
						{Name: "question"},
					}, nil
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
			err := e.Empty(out, tt.args.repos)
			if (err != nil) != tt.wantErr {
				t.Errorf("Empty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotOut := stripProgress(out.String())
			// For parallel execution, compare output lines in sorted order
			if sortedLines(gotOut) != nil && sortedLines(tt.wantOut) != nil {
				gotLines := sortedLines(gotOut)
				wantLines := sortedLines(tt.wantOut)
				if len(gotLines) != len(wantLines) {
					t.Errorf("Empty() gotOut = %q, want %q", gotOut, tt.wantOut)
				} else {
					for i := range gotLines {
						if gotLines[i] != wantLines[i] {
							t.Errorf("Empty() gotOut = %q, want %q", gotOut, tt.wantOut)
							break
						}
					}
				}
			}
			// Use order-independent comparison for API calls (parallel execution)
			if !containsAllListCalls(tt.mock.ListLabelsCalls, tt.wantListCall) {
				t.Errorf("Empty() wantListCall = %v, got %v", tt.wantListCall, tt.mock.ListLabelsCalls)
			}
			if !containsAllDeleteCalls(tt.mock.DeleteLabelCalls, tt.wantDeleteCall) {
				t.Errorf("Empty() wantDeleteCall = %v, got %v", tt.wantDeleteCall, tt.mock.DeleteLabelCalls)
			}
		})
	}
}
