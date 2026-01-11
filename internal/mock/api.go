package mock

import (
	"sync"

	"github.com/tnagatomi/gh-fuda/option"
)

type MockAPI struct {
	mu sync.Mutex

	CreateLabelFunc  func(label option.Label, repo option.Repo) error
	CreateLabelCalls []struct {
		Label option.Label
		Repo  option.Repo
	}

	UpdateLabelFunc  func(label option.Label, repo option.Repo) error
	UpdateLabelCalls []struct {
		Label option.Label
		Repo  option.Repo
	}

	DeleteLabelFunc  func(label string, repo option.Repo) error
	DeleteLabelCalls []struct {
		Label string
		Repo  option.Repo
	}

	ListLabelsFunc  func(repo option.Repo) ([]option.Label, error)
	ListLabelsCalls []struct {
		Repo option.Repo
	}

	GetRepositoryIDFunc  func(repo option.Repo) (option.GraphQLID, error)
	GetRepositoryIDCalls []struct {
		Repo option.Repo
	}

	GetLabelIDFunc  func(repo option.Repo, labelName string) (option.GraphQLID, error)
	GetLabelIDCalls []struct {
		Repo      option.Repo
		LabelName string
	}

	SearchLabelablesFunc  func(repo option.Repo, labelName string) ([]option.Labelable, error)
	SearchLabelablesCalls []struct {
		Repo      option.Repo
		LabelName string
	}

	AddLabelsToLabelableFunc  func(labelableID option.GraphQLID, labelIDs []option.GraphQLID) error
	AddLabelsToLabelableCalls []struct {
		LabelableID option.GraphQLID
		LabelIDs    []option.GraphQLID
	}

	RemoveLabelsFromLabelableFunc  func(labelableID option.GraphQLID, labelIDs []option.GraphQLID) error
	RemoveLabelsFromLabelableCalls []struct {
		LabelableID option.GraphQLID
		LabelIDs    []option.GraphQLID
	}
}

func (m *MockAPI) CreateLabel(label option.Label, repo option.Repo) error {
	m.mu.Lock()
	m.CreateLabelCalls = append(m.CreateLabelCalls, struct {
		Label option.Label
		Repo  option.Repo
	}{label, repo})
	m.mu.Unlock()

	if m.CreateLabelFunc != nil {
		return m.CreateLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) UpdateLabel(label option.Label, repo option.Repo) error {
	m.mu.Lock()
	m.UpdateLabelCalls = append(m.UpdateLabelCalls, struct {
		Label option.Label
		Repo  option.Repo
	}{label, repo})
	m.mu.Unlock()

	if m.UpdateLabelFunc != nil {
		return m.UpdateLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) DeleteLabel(label string, repo option.Repo) error {
	m.mu.Lock()
	m.DeleteLabelCalls = append(m.DeleteLabelCalls, struct {
		Label string
		Repo  option.Repo
	}{label, repo})
	m.mu.Unlock()

	if m.DeleteLabelFunc != nil {
		return m.DeleteLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) ListLabels(repo option.Repo) ([]option.Label, error) {
	m.mu.Lock()
	m.ListLabelsCalls = append(m.ListLabelsCalls, struct {
		Repo option.Repo
	}{repo})
	m.mu.Unlock()

	if m.ListLabelsFunc != nil {
		return m.ListLabelsFunc(repo)
	}

	return nil, nil
}

func (m *MockAPI) GetRepositoryID(repo option.Repo) (option.GraphQLID, error) {
	m.mu.Lock()
	m.GetRepositoryIDCalls = append(m.GetRepositoryIDCalls, struct {
		Repo option.Repo
	}{repo})
	m.mu.Unlock()

	if m.GetRepositoryIDFunc != nil {
		return m.GetRepositoryIDFunc(repo)
	}

	return "", nil
}

func (m *MockAPI) GetLabelID(repo option.Repo, labelName string) (option.GraphQLID, error) {
	m.mu.Lock()
	m.GetLabelIDCalls = append(m.GetLabelIDCalls, struct {
		Repo      option.Repo
		LabelName string
	}{repo, labelName})
	m.mu.Unlock()

	if m.GetLabelIDFunc != nil {
		return m.GetLabelIDFunc(repo, labelName)
	}

	return "", nil
}

func (m *MockAPI) SearchLabelables(repo option.Repo, labelName string) ([]option.Labelable, error) {
	m.mu.Lock()
	m.SearchLabelablesCalls = append(m.SearchLabelablesCalls, struct {
		Repo      option.Repo
		LabelName string
	}{repo, labelName})
	m.mu.Unlock()

	if m.SearchLabelablesFunc != nil {
		return m.SearchLabelablesFunc(repo, labelName)
	}

	return nil, nil
}

func (m *MockAPI) AddLabelsToLabelable(labelableID option.GraphQLID, labelIDs []option.GraphQLID) error {
	m.mu.Lock()
	m.AddLabelsToLabelableCalls = append(m.AddLabelsToLabelableCalls, struct {
		LabelableID option.GraphQLID
		LabelIDs    []option.GraphQLID
	}{labelableID, labelIDs})
	m.mu.Unlock()

	if m.AddLabelsToLabelableFunc != nil {
		return m.AddLabelsToLabelableFunc(labelableID, labelIDs)
	}

	return nil
}

func (m *MockAPI) RemoveLabelsFromLabelable(labelableID option.GraphQLID, labelIDs []option.GraphQLID) error {
	m.mu.Lock()
	m.RemoveLabelsFromLabelableCalls = append(m.RemoveLabelsFromLabelableCalls, struct {
		LabelableID option.GraphQLID
		LabelIDs    []option.GraphQLID
	}{labelableID, labelIDs})
	m.mu.Unlock()

	if m.RemoveLabelsFromLabelableFunc != nil {
		return m.RemoveLabelsFromLabelableFunc(labelableID, labelIDs)
	}

	return nil
}
