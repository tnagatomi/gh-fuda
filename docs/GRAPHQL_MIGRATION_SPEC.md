# GraphQL Migration Specification

## Overview

Migrate gh-fuda from GitHub REST API to GraphQL API in v3.0.0.

### Migration Goals

1. **Discussions Support**: REST API does not support label operations on Discussions, but GraphQL does
2. **Rate Limit Efficiency**: GraphQL is more efficient for bulk operations and fetching across multiple repositories

### Priority

Discussions support is the primary goal; rate limit efficiency is a secondary benefit.

---

## Technical Decisions

### API Client

- **Library**: Use `cli/go-gh` GraphQL client
  - Authentication is automatically integrated
  - Define queries using `shurcooL-graphql` style struct tags
- **Dependencies**: Completely remove `google/go-github`

### Interface Design

- **Approach**: Complete redesign
- **Interface Name**: Keep `APIClient` (no visible change from external perspective)

Complete APIClient interface after migration:

```go
type APIClient interface {
    // Repository label operations
    CreateLabel(label option.Label, repo option.Repo) error
    UpdateLabel(label option.Label, repo option.Repo) error
    DeleteLabel(label string, repo option.Repo) error
    ListLabels(repo option.Repo) ([]option.Label, error)

    // Labelable operations (new)
    AddLabelsToLabelable(labelableID string, labelIDs []string) error
    RemoveLabelsFromLabelable(labelableID string, labelIDs []string) error
    SearchLabelables(repo option.Repo, labelName string) ([]option.Labelable, error)

    // Helper methods (new)
    GetRepositoryID(repo option.Repo) (string, error)
    GetLabelID(repo option.Repo, labelName string) (string, error)
}
```

### Data Structures

New types for merge command:

```go
// In option/ package

// Labelable represents any GitHub resource that can have labels
type Labelable struct {
    ID     string // GraphQL node ID
    Number int
    Title  string
    Type   string // "Issue", "PullRequest", "Discussion"
}
```

### Query Management

- **Approach**: Hardcoded (query strings directly in Go code)
- **Struct Tags**: Adopt `shurcooL-graphql` format

### Pagination

- **Approach**: Simple implementation (loop until all items are fetched)
- **Cursor Management**: Use GraphQL cursor-based pagination

### ID Resolution Strategy

GraphQL mutations require node IDs instead of owner/repo or label names.

**Repository IDs:**
- Fetch repository node ID on first operation per repository
- Cache in memory for subsequent operations in the same command execution
- Query: `repository(owner: $owner, name: $name) { id }`

**Label IDs:**
- For merge command: fetch all labels at start to build name→ID map
- Validate target label exists before proceeding
- Error if source or target label not found

---

## Multiple Repository Operations

### Query Strategy

- **Approach**: Parallel queries (individual query per repository, executed in parallel)
  - Maintains type safety with struct tags
  - Simpler implementation and testing compared to batch queries with aliases
  - GraphQL rate limiting is based on query cost, not request count, so parallel queries are acceptable

### Parallel Execution

- **Approach**: Limited parallelization (parallel execution with goroutines)
- **Concurrency Limit**: 5 (fixed)
- **Output Format**:
  - During processing: Progress counter (X/Y completed)
  - After completion: Output all repository results together

### On Failure

- Record failure and continue with remaining operations (same as current behavior)

---

## Error Handling

### Error Types

- **Approach**: Create new error types (to handle GraphQL-specific error information)
- **Design**: Simple (message and error type only)
  - Do not include GraphQL details (location, path) in error types

### Error Mapping

Map GraphQL error types to custom error types:

| GraphQL Error Type | Custom Error Type | Description |
|-------------------|-------------------|-------------|
| `NOT_FOUND` | `NotFoundError` | Resource not found |
| `FORBIDDEN` | `ForbiddenError` | Access denied |
| `UNAUTHORIZED` | `UnauthorizedError` | Authentication required |
| `INSUFFICIENT_SCOPES` | `ScopeError` | Token lacks required scope |
| `RATE_LIMITED` | `RateLimitError` | Rate limit exceeded |
| `ALREADY_EXISTS` (422) | `AlreadyExistsError` | Label already exists |

Detection logic:
1. Check HTTP status code first (401, 403, 429)
2. Parse GraphQL `errors` array in response body
3. Match `type` field or error message patterns

### Partial Success

- **Approach**: Partial success (reflect successful parts, report only failures)

### Scope Errors

- **Approach**: Custom message explicitly stating "Required scope: write:discussion"
- Parse error message to detect scope-related errors

### Rate Limiting

- **Approach**: Error exit (display error and exit when rate limited)
- **Auto-wait**: None

---

## New Feature: merge Command

### Overview

Merge label A into label B. Replace label A with B on all Issue/PR/Discussion, then delete label A from the repository.

### CLI Interface

```
gh fuda merge -R owner/repo --from labelA --to labelB
```

### Behavior

1. Search for all labelables (Issue/PR/Discussion) with label A in the repository using search query
2. Remove label A and add label B on each labelable
3. Delete label A from the repository

### Options

- `--from`: Source label name (required)
- `--to`: Target label name (required)
- `-R, --repos`: Target repositories (required)
- `-y`: Skip confirmation prompt
- `--dry-run`: Show what would be changed without executing

### Confirmation Prompt

- Display "Will replace labels on X Issue/PR/Discussion" before execution
- Can be skipped with `-y` flag

### Discussion Support

- Use GraphQL `search` query with `type: DISCUSSION` to find discussions with specific labels
  - Direct `labels` filter is not available in discussions query
- Example query:
  ```graphql
  {
    search(
      type: DISCUSSION
      query: "repo:owner/repo label:\"labelA\""
      first: 100
    ) {
      nodes {
        ... on Discussion {
          id
          number
          title
        }
      }
    }
  }
  ```

### Detailed Workflow

1. **Initialization**
   - Fetch repository ID
   - Fetch all labels to build name→ID map
   - Validate source label (`--from`) exists
   - Validate target label (`--to`) exists

2. **Search Phase**
   - Search for issues with source label
   - Search for pull requests with source label
   - Search for discussions with source label
   - Collect all labelables with their node IDs

3. **Confirmation Phase** (if not `-y`)
   - Display: "Found X issues, Y pull requests, Z discussions with label 'labelA'"
   - Prompt: "Replace with 'labelB' and delete 'labelA'? (y/N)"
   - If declined, exit with no changes

4. **Replacement Phase**
   - For each labelable:
     - Remove source label (`removeLabelsFromLabelable`)
     - Add target label (`addLabelsToLabelable`)
   - Continue on individual failures, collect errors

5. **Cleanup Phase**
   - Delete source label from repository (`deleteLabel`)

6. **Summary**
   - Display: "Replaced labels on X issues, Y pull requests, Z discussions"
   - If errors occurred, display error count and details

### Error Handling for merge

- If search fails for any resource type (issues/PRs/discussions), report error but continue with others
- If label replacement fails for individual items, continue with remaining items
- Track success/failure counts per resource type
- Include in summary: "Processed X issues, Y PRs; failed to process Z discussions: [error]"

### Edge Cases for merge

**Target label B does not exist in repository:**
- Error and exit before any changes
- Message: "Error: label 'labelB' not found in repository"

**Labelable already has both label A and B:**
- Only remove label A (B is already present)
- Do not attempt to add B again
- Count as successful replacement in summary

---

## CLI Changes

### Breaking Changes (v3.0.0)

#### Confirmation Flag Changes

| Command | Before | After |
|---------|--------|-------|
| delete | `--force` | `-y` |
| empty | `--force` | `-y` |
| sync | `--force` | `-y` |
| merge | - | `-y` (new) |

**Note**: `create --force` means "overwrite existing labels" and is retained.

### Migration Path

- v3.0.0 supports both `--force` and `-y` for backward compatibility
- `--force` shows deprecation warning: "Warning: --force is deprecated, use -y instead"
- Future version may remove `--force` support

### CLI Compatibility

- Existing command names and basic flag names are maintained
- Only the confirmation flags above are changed

---

## Testing Strategy

### Approach

- **API Layer**: Maintain HTTP-level mocking with gock
  - GraphQL requests are also HTTP POST, so existing patterns can be reused
- **Executor Layer**: Continue with interface mocks

### GraphQL Mock Pattern Example

```go
func TestCreateLabel(t *testing.T) {
    defer gock.Off()

    gock.New("https://api.github.com").
        Post("/graphql").
        MatchType("json").
        JSON(map[string]interface{}{
            "query": `mutation CreateLabel($input: CreateLabelInput!) {
                createLabel(input: $input) {
                    label { id name color description }
                }
            }`,
            "variables": map[string]interface{}{
                "input": map[string]interface{}{
                    "repositoryId": "R_123",
                    "name":         "bug",
                    "color":        "ff0000",
                    "description":  "Something is broken",
                },
            },
        }).
        Reply(200).
        JSON(map[string]interface{}{
            "data": map[string]interface{}{
                "createLabel": map[string]interface{}{
                    "label": map[string]interface{}{
                        "id":          "LA_123",
                        "name":        "bug",
                        "color":       "ff0000",
                        "description": "Something is broken",
                    },
                },
            },
        })

    // Test implementation...
}
```

### Test Migration

- Full replacement (rewrite all tests for GraphQL)

---

## Environment Support

### GitHub Enterprise Server

- **Testing**: None (test only on github.com)
- **Support**: Address via user reports
- `cli/go-gh` automatically handles GH_HOST/GH_TOKEN

### REST API

- **Approach**: Complete replacement (fully deprecate REST)
- No fallback

---

## Output Format

### Progress Display

- During processing: Counter format (X/Y completed)
  - Fetch totalCount in advance for display
- After completion: Output all repository results together

### Output Formats

- v3.0.0: Text format only
- JSON output support will be considered in future versions

### Verbose Flag

- Not added (debugging is delegated to `gh --verbose`)

---

## Dry-run Mode

- **Approach**: Same as current (skip mutations and output messages only)
- All destructive commands support dry-run mode, including merge

---

## Release Plan

### Version

- Release as v3.0.0 (major release with breaking changes)

### Implementation Order

Phased implementation approach (single v3.0.0 release):

1. **Phase 1**: Basic API migration
   - Migrate CRUD operations (create, list, delete, update labels)
   - No parallel execution yet
   - Validate with tests

2. **Phase 2**: Parallel execution and optimizations
   - Add goroutine-based parallel execution
   - Implement progress counter
   - Add buffered output

3. **Phase 3**: merge command
   - Implement search query for labelables
   - Add label replacement logic
   - Implement confirmation prompt

### Documentation

- README update only (no migration guide)

---

## Implementation Notes

### GraphQL Mutations

Repository label operations:
- `createLabel`
- `updateLabel`
- `deleteLabel`

Labelable operations:
- `addLabelsToLabelable`
- `removeLabelsFromLabelable`

### Search Queries for Labelables

Use search query to find items with specific labels:

```graphql
# Issues and PRs
{
  search(type: ISSUE, query: "repo:owner/repo label:\"labelA\"", first: 100) {
    nodes {
      ... on Issue { id number title }
      ... on PullRequest { id number title }
    }
  }
}

# Discussions
{
  search(type: DISCUSSION, query: "repo:owner/repo label:\"labelA\"", first: 100) {
    nodes {
      ... on Discussion { id number title }
    }
  }
}
```

### Timeout

- Not needed (wait until completion)

### go-gh GraphQL Client Usage Example

```go
import "github.com/cli/go-gh/v2/pkg/api"

client, _ := api.NewGraphQLClient(api.ClientOptions{})

var query struct {
    Repository struct {
        Labels struct {
            Nodes []struct {
                Name        string
                Color       string
                Description string
            }
            PageInfo struct {
                HasNextPage bool
                EndCursor   string
            }
        } `graphql:"labels(first: 100, after: $cursor)"`
    } `graphql:"repository(owner: $owner, name: $name)"`
}

variables := map[string]interface{}{
    "owner":  graphql.String("owner"),
    "name":   graphql.String("repo"),
    "cursor": (*graphql.String)(nil),
}

err := client.Query("RepositoryLabels", &query, variables)
```