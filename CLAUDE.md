# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gh-fuda is a GitHub CLI extension that enables label operations across multiple repositories. It's written in Go and provides commands to list, create, delete, sync, and empty labels across GitHub repositories.

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
   - Each command (list, create, delete, sync, empty) has its own file
   - Commands parse arguments and delegate to executor
   - `create` and `sync` commands support JSON file input via `--json` flag
   - `create` command supports `--force` flag to update existing labels instead of failing
   - `list` command does not use dry-run mode as it's a read-only operation

2. **executor/** - Business logic layer
   - Contains the core functionality for label operations
   - Handles batch operations across multiple repositories
   - Supports dry-run mode for all operations (except `list` which is read-only)
   - Uses `ExecutionResult` to collect errors and provide operation summaries
   - Implements smart sync logic that diffs existing labels and updates only as needed
   - `List` method displays labels with their colors and descriptions

3. **api/** - GitHub API client wrapper
   - Defines `APIClient` interface for testability
   - Wraps google/go-github client
   - All API operations go through this layer
   - Provides custom error types (`NotFoundError`, `ForbiddenError`, etc.) with `ResourceType` enum for better error categorization
   - Error messages are simplified (e.g., "repository not found" instead of full details)
   - `ListLabels` supports pagination to handle repositories with more than 100 labels

4. **parser/** - Command-line option parsing
   - Handles file-based input for labels and repositories
   - Validates and transforms user input
   - `LabelFromJSON()` function to parse labels from JSON files

5. **option/** - Option structures
   - Defines data structures for command options
   - Shared across parser and executor layers

## Testing Approach

- Table-driven tests are the standard pattern
- Mock API client (`internal/mock/api.go`) for unit testing executor logic
- HTTP mocking with gock for API client tests
- Test files are colocated with implementation files

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
- Executor methods accept structured data instead of strings:
  - `List(out io.Writer, repos []option.Repo) error` - Lists all labels with their details
  - `Create(out io.Writer, repos []option.Repo, labels []option.Label, force bool) error` - Creates labels; with force=true, updates existing labels instead of failing
  - `Delete(out io.Writer, repos []option.Repo, labels []string) error`
  - `Sync(out io.Writer, repos []option.Repo, labels []option.Label) error`
  - `Empty(out io.Writer, repos []option.Repo) error`
- All executor functions accept APIClient interface for dependency injection

## Error Handling

- Operations continue even if individual label operations fail
- Errors are collected using `ExecutionResult` structure
- Summary is displayed at the end showing success/failure counts (format: "Summary: X repositories succeeded, Y failed")
- Exit code 1 if any operations failed
- Custom error types for common GitHub API errors (404, 403, etc.) with `ResourceType` enum to distinguish between repository and label errors
- Simplified error messages without redundant details (e.g., "repository not found" instead of "repository 'owner/repo' not found")
- Command usage is not displayed for runtime errors (only for argument/flag errors)

## File Input Support

The `create` and `sync` commands support both JSON and YAML file input for labels.

### JSON Format

```json
[
  {
    "name": "bug",
    "color": "d73a4a",
    "description": "Something isn't working"
  },
  {
    "name": "enhancement",
    "color": "a2eeef",
    "description": "New feature or request"
  }
]
```

### YAML Format

```yaml
- name: bug
  color: d73a4a
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
- `test.yml` - Runs tests on multiple OS (Ubuntu, Windows, macOS)
- `golangci-lint.yml` - Code quality checks
- `release.yml` - Automated release process

## Version History

- **v2.0.0** (upcoming) - Breaking change: `ListLabels` now returns `[]option.Label` instead of `[]string`. Added `list` command with pagination support.
- **v1.0.0** - Initial stable release
