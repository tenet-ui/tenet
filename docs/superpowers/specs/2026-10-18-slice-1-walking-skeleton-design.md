# Slice 1: Walking skeleton — design

Approved: 2026-10-18

Self-approved under the project's no-approval-gates rule (see CLAUDE.md, "Decide, don't ask").

## Goal

A thin, end-to-end path through tenet: a Go binary that runs as a stdio MCP server, exposes one tool, `get_point_of_view`, that returns `design-system/point-of-view.md` from a project directory, plus the repo scaffolding (tests, lint, format, CI, README) that every later slice builds on.

## Acceptance criteria (from docs/roadmap.md)

1. `go run ./cmd/tenet` launches a stdio MCP server. A client (Claude Code) lists one tool, `get_point_of_view`, which returns the text of `design-system/point-of-view.md` from the working directory.
2. The repo has a Go `.gitignore`, one passing `go test`, `gofmt` and `golangci-lint` check, CI running them on push, and a README with run instructions for Claude Code.

## Decisions

| # | Decision | Reason |
|---|---|---|
| D1 | Module path `github.com/tenet-ui/tenet`. | Matches the GitHub repository https://github.com/tenet-ui/tenet. |
| D2 | MCP SDK `github.com/modelcontextprotocol/go-sdk` at **v1.8.0** (latest stable tag). | ADR-0007. Pre-release tags are excluded. |
| D3 | Go toolchain: the `go` directive written by `go mod init` (1.27 locally). CI uses `actions/setup-go` with `go-version-file: go.mod`. | One version source; CI cannot drift from the module. |
| D4 | Lint: `golangci-lint` **v2.15.0**, run as `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`, both locally and in CI. | Pins the linter without a global install. The machine's global binary is v1.51.2, which cannot read a v2 config. |
| D5 | Lint config `.golangci.yml` (version "2") enables the default linters plus `gofmt` and `errcheck`. | Minimal set that catches the usual mistakes. |
| D6 | Format check: `test -z "$(gofmt -l .)"`. | Built into the Go toolchain; no extra dependency. |
| D7 | Project root = the process working directory at startup, overridable with `--dir <path>`. | The roadmap says "from the working directory". Claude Code may not set a cwd for stdio servers, so `--dir` is the reliable override. This is the one addition beyond the roadmap wording. |
| D8 | The point-of-view file is fixed at `design-system/point-of-view.md` under the root. The tool takes no input. | No caller-controlled path, so no path-traversal surface. |
| D9 | stdout carries only the MCP protocol. All logs go to stderr. | Writing anything else to stdout corrupts the stdio session. |
| D10 | Missing, empty, or directory-typed point-of-view file → a tool result with `IsError: true` and an actionable message. Not a JSON-RPC error. | The client should see a readable tool failure, not a protocol failure. |
| D11 | Empty file (zero bytes or only whitespace) counts as an error (`ErrPointOfViewEmpty`). | An empty point of view is useless to the AI and usually means a scaffolding mistake. |
| D12 | No file size limit and no encoding check in slice 1. Bytes are returned as text as-is. | YAGNI for slice 1; slice 3 owns response-size limits (roadmap: under 4 KB). |
| D13 | No `Makefile`. Commands live in the README and CI. | Fewer moving parts for the walking skeleton. |
| D14 | CI: one GitHub Actions workflow, `.github/workflows/ci.yml`, triggered on `push` and `pull_request`, job `check` running gofmt, golangci-lint and `go test ./...`. | Matches the acceptance criterion. Unverified until a remote exists and the workflow runs. |
| D15 | Tests use the SDK's in-memory transport (`mcp.NewInMemoryTransports`) to drive the real server in-process, plus plain unit tests for the file reader. No subprocess test. | Fast and deterministic. The stdio entry point is covered by a smoke test of `main` (see Testing). |

## Architecture

```
tenet/
├── cmd/tenet/main.go          # thin entry point: flags, root dir, run server on stdio
├── internal/designsystem/          # reads design-system files; no MCP types here
│   ├── pointofview.go              # ReadPointOfView
│   └── pointofview_test.go
├── internal/server/                # builds the MCP server and registers tools
│   ├── server.go                   # New(root) *mcp.Server
│   └── server_test.go              # in-memory client/server round trip
├── testdata/project/design-system/point-of-view.md   # fixture project for tests and README
├── .github/workflows/ci.yml
├── .golangci.yml
├── .gitignore
├── go.mod / go.sum
└── README.md
```

Dependency direction: `cmd` → `server` → `designsystem`. `designsystem` imports only the standard library, so slice 2 and 3 can grow it without touching MCP code.

### Module interfaces

**`internal/designsystem`**

```go
package designsystem

const PointOfViewPath = "design-system/point-of-view.md"

var (
    ErrPointOfViewMissing = errors.New("point of view file not found")
    ErrPointOfViewEmpty   = errors.New("point of view file is empty")
)

// ReadPointOfView returns the contents of root/design-system/point-of-view.md.
// It returns an error wrapping ErrPointOfViewMissing when the file does not
// exist, ErrPointOfViewEmpty when it holds only whitespace, and any other
// error (for example, a directory at that path) unwrapped.
func ReadPointOfView(root string) (string, error)
```

- A missing file is detected with `errors.Is(err, fs.ErrNotExist)` on the `os.ReadFile` error and re-wrapped as `ErrPointOfViewMissing`.
- The returned string is the file contents unchanged, including its trailing newline.

**`internal/server`**

```go
package server

const Version = "0.1.0"

// New returns an MCP server rooted at root, with the get_point_of_view tool registered.
func New(root string) *mcp.Server
```

Tool registration uses `mcp.AddTool` with an empty input struct:

- Name: `get_point_of_view`
- Description: "Returns the brand point of view of this project's design system (design-system/point-of-view.md)."
- Input: `struct{}` (schema: empty object)
- Output: `struct{}` (the text is returned in content, not as structured output)

Handler behaviour:

- Success → `CallToolResult` with one `TextContent` holding the file text, `IsError` false.
- `ErrPointOfViewMissing` → `CallToolResult` with `IsError: true` and text `design-system/point-of-view.md not found under <root>. Create it to describe the brand point of view.`
- `ErrPointOfViewEmpty` → `IsError: true`, text `design-system/point-of-view.md under <root> is empty. Write the point of view before using this tool.`
- Any other error → `IsError: true`, text `could not read design-system/point-of-view.md: <err>`.

The handler never returns a non-nil Go error, so the SDK never turns a file problem into a JSON-RPC error.

**`cmd/tenet`**

```
tenet [--dir <path>]
```

- `--dir` defaults to the process working directory (`os.Getwd`). The path is made absolute with `filepath.Abs`. If it does not exist or is not a directory, the process logs the problem to stderr and exits with status 2.
- Startup logs one line to stderr: `tenet <Version> serving <abs root> on stdio`.
- Runs `server.New(root).Run(ctx, &mcp.StdioTransport{})` with a context cancelled on SIGINT or SIGTERM (`signal.NotifyContext`). Exits 0 when the client closes stdin or the context is cancelled, 1 on any other error.

## Data flow

1. Client starts the process and sends `initialize`; the SDK answers.
2. Client sends `tools/list`; the server returns `get_point_of_view`.
3. Client sends `tools/call` for `get_point_of_view`; the handler calls `ReadPointOfView(root)` and returns the text or the `IsError` result.
4. The file is read on every call, so edits show up without restarting the server. No caching in slice 1.

## Error and edge cases

| Case | Behaviour |
|---|---|
| Point-of-view file missing | Tool result `IsError: true`, actionable message, process keeps running. |
| File exists but is empty or whitespace only | Tool result `IsError: true` (empty message). |
| `design-system` is a file, not a directory | Read fails with a non-ENOENT error → "could not read" message. |
| `design-system/point-of-view.md` is a directory | Read fails → "could not read" message. |
| `--dir` points to a missing path | Exit 2 with a message on stderr before serving. |
| `--dir` points to a regular file | Exit 2 with a message on stderr. |
| Relative `--dir` | Converted to an absolute path at startup. |
| Client sends a call with arguments to `get_point_of_view` | Arguments are ignored; the schema has no properties. |
| Client closes stdin | `Run` returns; process exits 0. |
| SIGINT or SIGTERM | Context cancelled; `Run` returns; exit 0. |
| File changes between calls | Next call returns the new contents (no cache). |
| Non-UTF-8 file | Returned as-is; no validation (D12). |

## Testing

**Unit (`internal/designsystem`)**, table-driven with `t.TempDir()`:
- returns contents and keeps the trailing newline
- missing `design-system/` → `ErrPointOfViewMissing`
- missing file inside an existing `design-system/` → `ErrPointOfViewMissing`
- empty file → `ErrPointOfViewEmpty`
- whitespace-only file → `ErrPointOfViewEmpty`
- `point-of-view.md` is a directory → error that is neither `ErrPointOfViewMissing` nor `ErrPointOfViewEmpty`

**Server (`internal/server`)**, in-memory transport against `testdata/project` and temp dirs:
- `tools/list` returns exactly one tool, named `get_point_of_view`, with an object input schema
- call on the fixture returns the fixture text, `IsError` false
- call with a temp root that lacks the file returns `IsError: true` and text containing `not found`
- call with an empty file returns `IsError: true` and text containing `empty`
- calling an unknown tool name returns the SDK's error (no crash)

**Entry point (`cmd/tenet`)**: a test of `parseRoot(args []string, cwd string) (string, error)`, a pure function holding the `--dir` logic, covering default, relative, missing path, and file-not-directory. `main` itself only wires flags to `server.New` and is not unit-tested; the manual check in the README covers the stdio path.

**Stdio check (acceptance criterion 1)**: the implementer runs the binary once, `go run ./cmd/tenet --dir testdata/project`, with piped JSON-RPC input (`initialize`, `notifications/initialized`, `tools/list`, `tools/call`), and confirms that stdout carries only JSON-RPC responses and that the call returns the fixture text. The README documents the same commands for a user to repeat. Connecting the real Claude Code client is recorded as not run in slice 1, because it needs an interactive client.

**Quality gates**, all of which must pass before the slice is committed:
- `go test ./...`
- `test -z "$(gofmt -l .)"`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`

## Repository files

- `.gitignore`: `/tenet`, `/dist/`, `*.test`, `*.out`, `.DS_Store`, `.idea/`, `.vscode/`. `.scratch/` is NOT ignored, because it holds the local issue tracker.
- `.golangci.yml`: `version: "2"`, default linters, plus `gofmt` and `errcheck` in the formatter and linter lists.
- `README.md`: what tenet is (one paragraph from GLOSSARY.md); requirements (Go toolchain matching `go.mod`); build and run; running the tests and quality gates; connecting Claude Code (`claude mcp add --transport stdio tenet -- /absolute/path/to/tenet --dir /absolute/path/to/project`); a note that the point-of-view file is the first thing to write.
- `testdata/project/design-system/point-of-view.md`: a short fixture point of view, written for this repo's tests, not a Stripe artifact.

## Out of scope for slice 1

Tokens, rules, components, templates, flows, `validate`, `init`, critique, the Stripe pack, packaging or release binaries, versioned tool schemas, and any caching or size limits.

## Open risks

- CI is unverified until a GitHub remote exists and the workflow runs. The plan includes a YAML lint check of the workflow locally, but it cannot prove the runner behaves.
- The golangci-lint `go run ...@v2.15.0` path downloads the linter on first use; it needs network access in CI.
