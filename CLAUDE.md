# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gh-fuda is a GitHub CLI extension that enables label operations across multiple repositories. It's written in Go and provides commands to create, delete, sync, and empty labels across GitHub repositories.

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
   - Each command (create, delete, sync, empty) has its own file
   - Commands parse arguments and delegate to executor

2. **executor/** - Business logic layer
   - Contains the core functionality for label operations
   - Handles batch operations across multiple repositories
   - Supports dry-run mode for all operations

3. **api/** - GitHub API client wrapper
   - Defines `APIClient` interface for testability
   - Wraps google/go-github client
   - All API operations go through this layer

4. **parser/** - Command-line option parsing
   - Handles file-based input for labels and repositories
   - Validates and transforms user input

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
- All executor functions accept this interface for dependency injection

## Error Handling

- Operations continue even if individual label operations fail
- Errors are collected and reported at the end
- Exit code 1 if any operations failed

## CI/CD

GitHub Actions workflows:
- `test.yml` - Runs tests on multiple OS (Ubuntu, Windows, macOS)
- `golangci-lint.yml` - Code quality checks
- `release.yml` - Automated release process