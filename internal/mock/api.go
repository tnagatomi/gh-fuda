package mock

import "github.com/tnagatomi/gh-fuda/option"

type MockAPI struct {
	CreateLabelFunc func(label option.Label, repo option.Repo) error
	CreateLabelCalls []struct {
		Label option.Label
		Repo  option.Repo
	}

	UpdateLabelFunc func(label option.Label, repo option.Repo) error
	UpdateLabelCalls []struct {
		Label option.Label
		Repo  option.Repo
	}

	DeleteLabelFunc func(label string, repo option.Repo) error
	DeleteLabelCalls []struct {
		Label string
		Repo  option.Repo
	}

	ListLabelsFunc func(repo option.Repo) ([]option.Label, error)
	ListLabelsCalls []struct {
		Repo option.Repo
	}

	GetRepositoryIDFunc func(repo option.Repo) (string, error)
	GetRepositoryIDCalls []struct {
		Repo option.Repo
	}

	GetLabelIDFunc func(repo option.Repo, labelName string) (string, error)
	GetLabelIDCalls []struct {
		Repo      option.Repo
		LabelName string
	}
}

func (m *MockAPI) CreateLabel(label option.Label, repo option.Repo) error {
	m.CreateLabelCalls = append(m.CreateLabelCalls, struct {
		Label option.Label
		Repo  option.Repo
	}{label, repo})

	if m.CreateLabelFunc != nil {
		return m.CreateLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) UpdateLabel(label option.Label, repo option.Repo) error {
	m.UpdateLabelCalls = append(m.UpdateLabelCalls, struct {
		Label option.Label
		Repo  option.Repo
	}{label, repo})

	if m.UpdateLabelFunc != nil {
		return m.UpdateLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) DeleteLabel(label string, repo option.Repo) error {
	m.DeleteLabelCalls = append(m.DeleteLabelCalls, struct {
		Label string
		Repo  option.Repo
	}{label, repo})

	if m.DeleteLabelFunc != nil {
		return m.DeleteLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) ListLabels(repo option.Repo) ([]option.Label, error) {
	m.ListLabelsCalls = append(m.ListLabelsCalls, struct {
		Repo option.Repo
	}{repo})

	if m.ListLabelsFunc != nil {
		return m.ListLabelsFunc(repo)
	}

	return nil, nil
}

func (m *MockAPI) GetRepositoryID(repo option.Repo) (string, error) {
	m.GetRepositoryIDCalls = append(m.GetRepositoryIDCalls, struct {
		Repo option.Repo
	}{repo})

	if m.GetRepositoryIDFunc != nil {
		return m.GetRepositoryIDFunc(repo)
	}

	return "", nil
}

func (m *MockAPI) GetLabelID(repo option.Repo, labelName string) (string, error) {
	m.GetLabelIDCalls = append(m.GetLabelIDCalls, struct {
		Repo      option.Repo
		LabelName string
	}{repo, labelName})

	if m.GetLabelIDFunc != nil {
		return m.GetLabelIDFunc(repo, labelName)
	}

	return "", nil
}
