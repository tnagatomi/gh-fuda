# AGENTS.md

## Build & Test Commands
- **Build**: `go build`
- **Test all**: `go test ./...`
- **Test single package**: `go test ./path/to/package`
- **Test specific function**: `go test -run TestFunctionName ./path/to/package`
- **Coverage**: `go test -cover ./...`
- **Lint**: `golangci-lint run` (install golangci-lint first)

## Code Style
- Use standard Go imports order: stdlib, third-party, local packages
- Follow layered architecture: cmd → executor → client/parser → option
- Use table-driven tests with struct slices and t.Run() subtests
- Interfaces for testability (APIClient, mock in internal/mock/)
- Error wrapping with fmt.Errorf("context: %v", err)
- Simplified error messages without redundant details
- Custom error types with ResourceType enum for API errors
- Collect errors with ExecutionResult, continue operations on failures
- MIT license header on all .go files
- Use structured option types instead of raw strings
- Cobra commands with dependency injection for testing