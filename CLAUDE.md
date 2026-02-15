# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gh-fuda is a GitHub CLI extension that enables label operations across multiple repositories. It's written in Go and provides commands to list, create, delete, sync, empty, and merge labels across GitHub repositories.

## Commands

### Development
- **Build**: `go build`
- **Run tests**: `go test ./...`
- **Run specific test**: `go test -run TestName ./path/to/package`
- **Coverage**: `go test -cover ./...`
- **Lint locally**: Install golangci-lint and run `golangci-lint run`

### Installation
- **Install locally**: `go install`
- **Install via gh**: `gh extension install tnagatomi/gh-fuda`

## Architecture

The codebase follows a clean layered architecture:

1. **cmd/** - CLI command implementations using Cobra
   - Each command (list, create, delete, sync, empty, merge) has its own file
   - Commands parse arguments and delegate to executor
   - `create` and `sync` commands support JSON file input via `--json` flag
   - `create` command supports `--force` flag to update existing labels instead of failing
   - `list` command does not use dry-run mode as it's a read-only operation
   - `merge` command merges a source label into a target label across issues, PRs, and discussions

2. **executor/** - Business logic layer
   - Contains the core functionality for label operations
   - Handles batch operations across multiple repositories
   - Supports dry-run mode for all operations (except `list` which is read-only)
   - Uses `ExecutionResult` to collect errors and provide operation summaries
   - Implements smart sync logic that diffs existing labels and updates only as needed
   - `List` method displays labels with their colors and descriptions

3. **api/** - GitHub API client wrapper
   - Defines `APIClient` interface for testability
   - Uses GitHub GraphQL API via cli/go-gh v2 package
   - All API operations go through this layer
   - Provides custom error types (`NotFoundError`, `ForbiddenError`, `ScopeError`, etc.) with `ResourceType` enum for better error categorization
   - Error messages are simplified (e.g., "repository not found" instead of full details)
   - `ListLabels` supports cursor-based pagination to handle repositories with more than 100 labels

4. **parser/** - Command-line option parsing
   - Handles file-based input for labels and repositories
   - Validates and transforms user input
   - `LabelFromJSON()` function to parse labels from JSON files
   - `GenerateColor()` function generates deterministic colors from label names using SHA-256 hash

5. **option/** - Option structures
   - Defines data structures for command options
   - Shared across parser and executor layers

## Testing Approach

### Unit Tests
- Table-driven tests are the standard pattern
- Mock API client (`internal/mock/api.go`) for unit testing executor logic
- HTTP mocking with gock for API client tests
- Test files are colocated with implementation files
- Run with: `go test ./...`

### E2E Tests
- Located in `e2e_test.go` at the project root
- Uses build tag `//go:build e2e` to separate from unit tests
- Executes actual CLI commands against real GitHub API
- Requires `GH_TOKEN` or `GITHUB_TOKEN` environment variable
- Test repositories configurable via `GH_FUDA_TEST_REPO_1` and `GH_FUDA_TEST_REPO_2` env vars (defaults: `tnagatomi/gh-fuda-test-1`, `tnagatomi/gh-fuda-test-2`)
- Run with: `go test -tags=e2e -v`
- Local-only (no CI workflow); run manually for validation

Example test pattern:
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        // input fields
        want    expectedType
        wantErr bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Key Interfaces

- `api.APIClient` - Main interface for GitHub API operations
  - `CreateLabel(label option.Label, repo option.Repo) error`
  - `UpdateLabel(label option.Label, repo option.Repo) error`
  - `DeleteLabel(label string, repo option.Repo) error`
  - `ListLabels(repo option.Repo) ([]option.Label, error)` - Returns full label details including color and description
  - `SearchLabelables(repo option.Repo, labelName string) ([]option.Labelable, error)` - Searches issues, PRs, and discussions with a label
  - `AddLabelsToLabelable(labelableID string, labelIDs []string) error` - Adds labels to an issue/PR/discussion
  - `RemoveLabelsFromLabelable(labelableID string, labelIDs []string) error` - Removes labels from an issue/PR/discussion
  - `GetRepositoryID(repo option.Repo) (string, error)` - Gets repository node ID for GraphQL operations
  - `GetLabelID(repo option.Repo, labelName string) (string, error)` - Gets label node ID for GraphQL operations
- Executor methods accept structured data instead of strings:
  - `List(out io.Writer, repos []option.Repo) error` - Lists all labels with their details
  - `Create(out io.Writer, repos []option.Repo, labels []option.Label, force bool) error` - Creates labels; with force=true, updates existing labels instead of failing
  - `Delete(out io.Writer, repos []option.Repo, labels []string) error`
  - `Sync(out io.Writer, repos []option.Repo, labels []option.Label) error`
  - `Empty(out io.Writer, repos []option.Repo) error`
  - `Merge(out io.Writer, repos []option.Repo, fromLabel, toLabel string) error` - Merges source label into target label
- All executor functions accept APIClient interface for dependency injection

## Error Handling

- Operations continue even if individual label operations fail
- Errors are collected using `ExecutionResult` structure
- Summary is displayed at the end showing success/failure counts (format: "Summary: X repositories succeeded, Y failed")
- Exit code 1 if any operations failed
- Custom error types for common GitHub API errors (404, 403, etc.) with `ResourceType` enum to distinguish between repository and label errors
- Simplified error messages without redundant details (e.g., "repository not found" instead of "repository 'owner/repo' not found")
- Command usage is not displayed for runtime errors (only for argument/flag errors)

## Label Input Support

The `create` and `sync` commands support multiple input formats for labels.

### Color Auto-Generation

When the color is omitted or empty, a color is automatically generated from the label name using a SHA-256 hash. This ensures that the same label name always produces the same color across all repositories.

### CLI Format (`--labels` / `-l`)

Supported formats:
- `name` - Name only, color auto-generated
- `name:color` - Name and color
- `name:color:description` - Name, color, and description
- `name::description` - Name and description, color auto-generated

Examples:
```bash
gh fuda create -R owner/repo -l "bug"
gh fuda create -R owner/repo -l "bug:d73a4a:Something isn't working"
gh fuda create -R owner/repo -l "bug::Something isn't working"
gh fuda create -R owner/repo -l "bug,enhancement:a2eeef,feature::New feature"
```

### JSON Format (`--json`)

The `color` field is optional. If omitted or empty, color is auto-generated.

```json
[
  {
    "name": "bug",
    "description": "Something isn't working"
  },
  {
    "name": "enhancement",
    "color": "a2eeef",
    "description": "New feature or request"
  }
]
```

### YAML Format (`--yaml`)

The `color` field is optional. If omitted or empty, color is auto-generated.

```yaml
- name: bug
  description: Something isn't working
- name: enhancement
  color: a2eeef
  description: New feature or request
```

### Usage

- `gh fuda create -R owner/repo --json labels.json`
- `gh fuda create -R owner/repo --yaml labels.yaml`
- `gh fuda sync -R owner/repo --json labels.json`
- `gh fuda sync -R owner/repo --yaml labels.yaml`

Note: `--json`, `--yaml`, and `-l/--labels` flags are mutually exclusive. You must use exactly one of these options.

## CI/CD

GitHub Actions workflows:
- `test.yml` - Runs unit tests on multiple OS (Ubuntu, Windows, macOS)
- `golangci-lint.yml` - Code quality checks
- `release.yml` - Automated release process

