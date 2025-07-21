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

	ListLabelsFunc func(repo option.Repo) ([]string, error)
	ListLabelsCalls []struct {
		Repo option.Repo
	}
}

func (m *MockAPI) CreateLabel(label option.Label, repo option.Repo) error {
	m.CreateLabelCalls = append(m.CreateLabelCalls, struct {
		Label option.Label
		Repo  option.Repo
	}{label,repo})

	if m.CreateLabelFunc != nil {
		return m.CreateLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) UpdateLabel(label option.Label, repo option.Repo) error {
	m.UpdateLabelCalls = append(m.UpdateLabelCalls, struct {
		Label option.Label
		Repo  option.Repo
	}{label,repo})

	if m.UpdateLabelFunc != nil {
		return m.UpdateLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) DeleteLabel(label string, repo option.Repo) error {
	m.DeleteLabelCalls = append(m.DeleteLabelCalls, struct {
		Label string
		Repo  option.Repo
	}{label,repo})

	if m.DeleteLabelFunc != nil {
		return m.DeleteLabelFunc(label, repo)
	}

	return nil
}

func (m *MockAPI) ListLabels(repo option.Repo) ([]string, error) {
	m.ListLabelsCalls = append(m.ListLabelsCalls, struct {
		Repo option.Repo
	}{repo})

	if m.ListLabelsFunc != nil {
		return m.ListLabelsFunc(repo)
	}

	return nil, nil
}
