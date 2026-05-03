# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project Overview

gh-fuda is a GitHub CLI extension (Go) for label operations across multiple repositories. Commands: `list`, `create`, `delete`, `sync`, `empty`, `merge`.

## Commands

- Build: `go build`
- Unit tests: `go test ./...`
- Coverage: `go test -cover ./...`
- Lint: `golangci-lint run`
- E2E tests: `go test -tags=e2e -v` (local only, requires `GH_TOKEN`/`GITHUB_TOKEN`; test repos via `GH_FUDA_TEST_REPO_1`/`_2`, defaults `tnagatomi/gh-fuda-test-{1,2}`)

## Architecture

Layered: `cmd/` (Cobra CLI) → `executor/` (business logic) → `api/` (GitHub GraphQL via `cli/go-gh` v2). Supporting: `parser/` (CLI/JSON/YAML input parsing), `option/` (shared structs).

- All operations except `list` support dry-run.
- Executors continue past per-repo failures and aggregate via `ExecutionResult` (exit 1 if any failed).
- `api.APIClient` is an interface for DI; mock at `internal/mock/api.go`.
- `sync` diffs existing labels and only updates what changed.
- `merge` rewrites labels on issues, PRs, and discussions (requires labelable IDs via GraphQL).
- `create --force` updates existing labels instead of erroring.
- `ListLabels` paginates via cursor (>100 labels).

## Testing

- Table-driven tests, colocated with implementation.
- API client tests mock HTTP with `gock`; executor tests use the mock APIClient.
- E2E tests live in `e2e_test.go` behind `//go:build e2e`.

## Error Handling

- Custom typed errors in `api/`: `NotFoundError`, `ForbiddenError`, `ScopeError`, `RateLimitError`, `TransientError`, with a `ResourceType` enum (repository vs label).
- Cobra `SilenceUsage` is on for runtime errors; usage prints only for arg/flag errors.

## Retry Behavior (`api/retry.go`)

All GraphQL calls go through `withRetry`. Default: 3 attempts, exponential backoff 1s→2s→4s (cap 8s). `sleep` is injectable for tests.

- Idempotent ops (queries, `UpdateLabel`, `AddLabelsToLabelable`, `RemoveLabelsFromLabelable`) retry on both `RateLimitError` and `TransientError`.
- Non-idempotent ops (`CreateLabel`, `DeleteLabel`) retry **only** on `RateLimitError`. Retrying these on transient errors risks observing `AlreadyExists`/`NotFound` from a committed prior call and reporting a false failure; 429s are gateway-rejected before the data layer, so they're safe.
- `net.Error` (timeouts, reset, DNS, deadline) is mapped to `TransientError` in `wrapGraphQLError`.
- `shurcooL-graphql` (used by go-gh's `Query`/`Mutate`) does not surface `*api.HTTPError` for non-2xx; `parseShurcoolStatusCode` recovers the status from the plain error message so 429/5xx still map to typed errors.

## Label Input

`create` and `sync` accept labels via `-l/--labels`, `--json`, or `--yaml` (mutually exclusive).

CLI format: `name`, `name:color`, `name:color:description`, or `name::description` (comma-separated for multiple). JSON/YAML use a list of `{name, color?, description?}`.

When `color` is empty/omitted, it's deterministically generated from the label name via SHA-256 (`parser.GenerateColor`), so the same name yields the same color across repos.

## CI

`.github/workflows/`: `test.yml` (Ubuntu/Windows/macOS), `golangci-lint.yml`, `release.yml`.
