# Slice 1: Walking Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A Go binary that runs as a stdio MCP server exposing one tool, `get_point_of_view`, which returns `design-system/point-of-view.md` from a project directory, with tests, lint, format, CI, and a README.

**Architecture:** `cmd/tenet` (flags and stdio entry point) → `internal/server` (MCP server and tool registration) → `internal/designsystem` (standard-library file reader with no MCP types). Tests use the SDK's in-memory transport for the server and table tests for the reader and flag parsing.

**Tech Stack:** Go (toolchain from `go.mod`), `github.com/modelcontextprotocol/go-sdk` v1.8.0, golangci-lint v2.15.0 run via `go run`, GitHub Actions.

**Spec:** [docs/superpowers/specs/2026-10-18-slice-1-walking-skeleton-design.md](../specs/2026-10-18-slice-1-walking-skeleton-design.md)

## Global Constraints

- Module path: `github.com/tenet-ui/tenet`
- MCP SDK: `github.com/modelcontextprotocol/go-sdk` at exactly `v1.8.0`
- Linter: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`
- Format: `test -z "$(gofmt -l .)"`
- Tests: `go test ./...`
- Point-of-view path (relative to root): `design-system/point-of-view.md`
- Tool name: `get_point_of_view`; server name `tenet`; `Version = "0.1.0"`
- stdout carries only MCP protocol; all logs go to stderr
- Do not commit anywhere in this plan; leave changes in the working tree

## Review Focus

- Point-of-view file missing → tool result with `IsError: true` and an actionable message; the process keeps serving (Task 2, `TestCallMissingFileIsToolError`).
- Point-of-view file empty or whitespace-only → `IsError: true` (Task 1, `TestReadPointOfViewEmptyFile`, `TestReadPointOfViewWhitespaceOnlyFile`; Task 2, `TestCallEmptyFileIsToolError`).
- `design-system` is a regular file, not a directory → generic read error, not reported as "not found" (Task 1, `TestReadPointOfViewDesignSystemIsFile`).
- `--dir` points to a missing path or a regular file → exit 2 before serving (Task 3, `TestParseRootMissingPath`, `TestParseRootFileNotDirectory`).
- Relative `--dir` resolves against the working directory, not the process root (Task 3, `TestParseRootRelativeDir`).

---

### Task 1: Module scaffold, quality config, and the point-of-view reader

**Model:** sonnet

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `.golangci.yml`
- Create: `testdata/project/design-system/point-of-view.md`
- Create: `internal/designsystem/pointofview.go`
- Test: `internal/designsystem/pointofview_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces:
  - package `designsystem` (import path `github.com/tenet-ui/tenet/internal/designsystem`)
  - `const PointOfViewPath = "design-system/point-of-view.md"`
  - `var ErrPointOfViewMissing error`, `var ErrPointOfViewEmpty error`
  - `func ReadPointOfView(root string) (string, error)`: returns the file contents unchanged (trailing newline kept). Wraps `ErrPointOfViewMissing` when the file or its directory does not exist; wraps `ErrPointOfViewEmpty` when the trimmed contents are empty; any other read error (for example a directory at the path, or `design-system` being a regular file) is returned unwrapped.

- [ ] **Step 1: Initialise the module and write the scaffold files**

Run:
```bash
go mod init github.com/tenet-ui/tenet
```

Create `.gitignore` with exactly this content:
```
/tenet
/dist/
*.test
*.out
.DS_Store
.idea/
.vscode/
```

Create `.golangci.yml` with exactly this content:
```yaml
version: "2"
linters:
  default: standard
formatters:
  enable:
    - gofmt
```

Create `testdata/project/design-system/point-of-view.md` with exactly this content:
```markdown
# Point of view

This fixture describes a calm, precise brand for tests. It is not a real product's point of view.
```

- [ ] **Step 2: Write the failing test**

Create `internal/designsystem/pointofview_test.go` with exactly this content:
```go
package designsystem

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writePointOfView(t *testing.T, root string, content string) {
	t.Helper()
	dir := filepath.Join(root, "design-system")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "point-of-view.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadPointOfViewReturnsContents(t *testing.T) {
	root := t.TempDir()
	writePointOfView(t, root, "# Point of view\n\nWe are calm.\n")

	got, err := ReadPointOfView(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "# Point of view\n\nWe are calm.\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReadPointOfViewMissingDesignSystemDir(t *testing.T) {
	_, err := ReadPointOfView(t.TempDir())
	if !errors.Is(err, ErrPointOfViewMissing) {
		t.Fatalf("got %v, want ErrPointOfViewMissing", err)
	}
}

func TestReadPointOfViewMissingFileInExistingDir(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "design-system"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := ReadPointOfView(root)
	if !errors.Is(err, ErrPointOfViewMissing) {
		t.Fatalf("got %v, want ErrPointOfViewMissing", err)
	}
}

func TestReadPointOfViewEmptyFile(t *testing.T) {
	root := t.TempDir()
	writePointOfView(t, root, "")

	_, err := ReadPointOfView(root)
	if !errors.Is(err, ErrPointOfViewEmpty) {
		t.Fatalf("got %v, want ErrPointOfViewEmpty", err)
	}
}

func TestReadPointOfViewWhitespaceOnlyFile(t *testing.T) {
	root := t.TempDir()
	writePointOfView(t, root, "  \n\t\n")

	_, err := ReadPointOfView(root)
	if !errors.Is(err, ErrPointOfViewEmpty) {
		t.Fatalf("got %v, want ErrPointOfViewEmpty", err)
	}
}

func TestReadPointOfViewDirectoryAtPath(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "design-system", "point-of-view.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := ReadPointOfView(root)
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrPointOfViewMissing) || errors.Is(err, ErrPointOfViewEmpty) {
		t.Fatalf("got %v, want a generic read error", err)
	}
}

func TestReadPointOfViewDesignSystemIsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "design-system"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadPointOfView(root)
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrPointOfViewMissing) || errors.Is(err, ErrPointOfViewEmpty) {
		t.Fatalf("got %v, want a generic read error", err)
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/designsystem/ -v`
Expected: FAIL with `undefined: ReadPointOfView` (and `undefined: ErrPointOfViewMissing`).

- [ ] **Step 4: Write the minimal implementation**

Create `internal/designsystem/pointofview.go` with exactly this content:
```go
package designsystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PointOfViewPath is the location of the point of view, relative to the project root.
const PointOfViewPath = "design-system/point-of-view.md"

var (
	// ErrPointOfViewMissing reports that the point of view file does not exist.
	ErrPointOfViewMissing = errors.New("point of view file not found")
	// ErrPointOfViewEmpty reports that the point of view file holds only whitespace.
	ErrPointOfViewEmpty = errors.New("point of view file is empty")
)

// ReadPointOfView returns the contents of root/design-system/point-of-view.md.
// The contents are returned unchanged, including any trailing newline.
func ReadPointOfView(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, PointOfViewPath))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("%w: %s", ErrPointOfViewMissing, PointOfViewPath)
		}
		return "", err
	}
	text := string(data)
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("%w: %s", ErrPointOfViewEmpty, PointOfViewPath)
	}
	return text, nil
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/designsystem/ -v`
Expected: PASS for all seven tests.

Then run: `gofmt -l internal/ && go vet ./internal/...`
Expected: no output from either command.

---

### Task 2: MCP server with the get_point_of_view tool

**Model:** sonnet

**Files:**
- Create: `internal/server/server.go`
- Test: `internal/server/server_test.go`
- Modify: `go.mod`, `go.sum` (via `go get`)

**Interfaces:**
- Consumes (from Task 1, package `designsystem`):
  - `func ReadPointOfView(root string) (string, error)`
  - `const PointOfViewPath`, `var ErrPointOfViewMissing`, `var ErrPointOfViewEmpty`
- Produces:
  - package `server` (import path `github.com/tenet-ui/tenet/internal/server`)
  - `const Version = "0.1.0"`
  - `func New(root string) *mcp.Server`: a server named `tenet` with one tool, `get_point_of_view`, with an empty object input schema and no structured output.
  - Tool result contract: success → one `*mcp.TextContent` with the file text, `IsError` false. Any failure → `IsError: true` with one `*mcp.TextContent`; the handler never returns a Go error.

- [ ] **Step 1: Add the SDK dependency**

Run:
```bash
go get github.com/modelcontextprotocol/go-sdk@v1.8.0
go mod tidy
```
Expected: `go.mod` lists `github.com/modelcontextprotocol/go-sdk v1.8.0`.

- [ ] **Step 2: Write the failing test**

Create `internal/server/server_test.go` with exactly this content:
```go
package server

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const fixtureRoot = "../../testdata/project"

func connect(t *testing.T, root string) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := New(root).Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func callPointOfView(t *testing.T, session *mcp.ClientSession) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_point_of_view",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call get_point_of_view: %v", err)
	}
	return res
}

func textOf(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) != 1 {
		t.Fatalf("got %d content items, want 1", len(res.Content))
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("got content %T, want *mcp.TextContent", res.Content[0])
	}
	return tc.Text
}

func TestListsOnlyGetPointOfView(t *testing.T) {
	session := connect(t, fixtureRoot)

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(res.Tools) != 1 {
		t.Fatalf("got %d tools, want 1", len(res.Tools))
	}
	if res.Tools[0].Name != "get_point_of_view" {
		t.Fatalf("got tool %q, want get_point_of_view", res.Tools[0].Name)
	}
}

func TestCallReturnsFixtureText(t *testing.T) {
	session := connect(t, fixtureRoot)
	want, err := os.ReadFile(filepath.Join(fixtureRoot, "design-system", "point-of-view.md"))
	if err != nil {
		t.Fatal(err)
	}

	res := callPointOfView(t, session)
	if res.IsError {
		t.Fatalf("got IsError, text: %s", textOf(t, res))
	}
	if got := textOf(t, res); got != string(want) {
		t.Fatalf("got %q, want %q", got, string(want))
	}
}

func TestCallMissingFileIsToolError(t *testing.T) {
	session := connect(t, t.TempDir())

	res := callPointOfView(t, session)
	if !res.IsError {
		t.Fatal("want IsError true")
	}
	if got := textOf(t, res); !strings.Contains(got, "not found") {
		t.Fatalf("got %q, want text containing \"not found\"", got)
	}
}

func TestCallEmptyFileIsToolError(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "design-system")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "point-of-view.md"), []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)

	res := callPointOfView(t, session)
	if !res.IsError {
		t.Fatal("want IsError true")
	}
	if got := textOf(t, res); !strings.Contains(got, "empty") {
		t.Fatalf("got %q, want text containing \"empty\"", got)
	}
}

func TestUnknownToolIsNotASuccess(t *testing.T) {
	session := connect(t, fixtureRoot)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "nope", Arguments: map[string]any{}})
	if err == nil && (res == nil || !res.IsError) {
		t.Fatal("want an error or IsError for an unknown tool")
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/server/ -v`
Expected: FAIL with `undefined: New`.

- [ ] **Step 4: Write the minimal implementation**

Create `internal/server/server.go` with exactly this content:
```go
package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/tenet-ui/tenet/internal/designsystem"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version is the server version reported to MCP clients.
const Version = "0.1.0"

// GetPointOfViewInput is the input of get_point_of_view. The tool takes no arguments.
type GetPointOfViewInput struct{}

// New returns an MCP server rooted at root, with the get_point_of_view tool registered.
func New(root string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "tenet", Version: Version}, nil)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_point_of_view",
		Description: "Returns the brand point of view of this project's design system (design-system/point-of-view.md).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ GetPointOfViewInput) (*mcp.CallToolResult, any, error) {
		return getPointOfView(root), nil, nil
	})
	return s
}

// getPointOfView never returns a Go error: file problems become IsError results
// so that the client sees a readable tool failure, not a protocol failure.
func getPointOfView(root string) *mcp.CallToolResult {
	text, err := designsystem.ReadPointOfView(root)
	if err != nil {
		return errorResult(err, root)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

func errorResult(err error, root string) *mcp.CallToolResult {
	var msg string
	switch {
	case errors.Is(err, designsystem.ErrPointOfViewMissing):
		msg = fmt.Sprintf("%s not found under %s. Create it to describe the brand point of view.", designsystem.PointOfViewPath, root)
	case errors.Is(err, designsystem.ErrPointOfViewEmpty):
		msg = fmt.Sprintf("%s under %s is empty. Write the point of view before using this tool.", designsystem.PointOfViewPath, root)
	default:
		msg = fmt.Sprintf("could not read %s: %v", designsystem.PointOfViewPath, err)
	}
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: msg}}}
}
```

If the SDK rejects `any` as the output type or the handler signature at compile time, the implementer must report BLOCKED with the compiler message rather than change the tool contract.

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/server/ -v`
Expected: PASS for all five tests.

Then run: `gofmt -l internal/ && go vet ./internal/...`
Expected: no output from either command.

---

### Task 3: Entry point with --dir handling

**Model:** haiku

**Files:**
- Create: `cmd/tenet/main.go`
- Test: `cmd/tenet/main_test.go`

**Interfaces:**
- Consumes (from Task 2, package `server`):
  - `const Version = "0.1.0"`
  - `func New(root string) *mcp.Server`
- Produces:
  - `func parseRoot(args []string, cwd string) (string, error)`: parses `--dir`; if empty, uses `cwd`; if relative, joins it to `cwd`; cleans the path; returns an error if the path does not exist (`--dir <path>: ...`) or is not a directory (`--dir <path> is not a directory`).
  - `func run(args []string, cwd string, stderr io.Writer) int`: exit codes 0 on normal stop, 1 on server error, 2 on bad `--dir`.
  - `func main()`: passes `os.Args[1:]` and `os.Getwd()` to `run`.

- [ ] **Step 1: Write the failing test**

Create `cmd/tenet/main_test.go` with exactly this content:
```go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRootDefaultsToCwd(t *testing.T) {
	cwd := t.TempDir()

	got, err := parseRoot(nil, cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != cwd {
		t.Fatalf("got %q, want %q", got, cwd)
	}
}

func TestParseRootAbsoluteDir(t *testing.T) {
	cwd := t.TempDir()
	dir := t.TempDir()

	got, err := parseRoot([]string{"--dir", dir}, cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != filepath.Clean(dir) {
		t.Fatalf("got %q, want %q", got, filepath.Clean(dir))
	}
}

func TestParseRootRelativeDir(t *testing.T) {
	cwd := t.TempDir()
	if err := os.Mkdir(filepath.Join(cwd, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := parseRoot([]string{"--dir", "sub"}, cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := filepath.Join(cwd, "sub"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseRootMissingPath(t *testing.T) {
	cwd := t.TempDir()

	_, err := parseRoot([]string{"--dir", "does-not-exist"}, cwd)
	if err == nil {
		t.Fatal("expected an error for a missing path")
	}
}

func TestParseRootFileNotDirectory(t *testing.T) {
	cwd := t.TempDir()
	file := filepath.Join(cwd, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := parseRoot([]string{"--dir", file}, cwd)
	if err == nil {
		t.Fatal("expected an error when --dir is a file")
	}
}

func TestParseRootUnknownFlag(t *testing.T) {
	_, err := parseRoot([]string{"--nope"}, t.TempDir())
	if err == nil {
		t.Fatal("expected an error for an unknown flag")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/tenet/ -v`
Expected: FAIL with `undefined: parseRoot`.

- [ ] **Step 3: Write the minimal implementation**

Create `cmd/tenet/main.go` with exactly this content:
```go
// Command tenet serves a project's design system to MCP clients over stdio.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/tenet-ui/tenet/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "tenet:", err)
		os.Exit(1)
	}
	os.Exit(run(os.Args[1:], cwd, os.Stderr))
}

func run(args []string, cwd string, stderr io.Writer) int {
	logger := log.New(stderr, "tenet: ", 0)
	root, err := parseRoot(args, cwd)
	if err != nil {
		logger.Println(err)
		return 2
	}
	logger.Printf("%s serving %s on stdio", server.Version, root)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.New(root).Run(ctx, &mcp.StdioTransport{}); err != nil {
		logger.Println(err)
		return 1
	}
	return 0
}

// parseRoot reads --dir and returns the cleaned absolute project root.
// It defaults to cwd, and a relative --dir is resolved against cwd.
func parseRoot(args []string, cwd string) (string, error) {
	fs := flag.NewFlagSet("tenet", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dir := fs.String("dir", "", "project root (default: current directory)")
	if err := fs.Parse(args); err != nil {
		return "", err
	}

	root := *dir
	if root == "" {
		root = cwd
	} else if !filepath.IsAbs(root) {
		root = filepath.Join(cwd, root)
	}
	root = filepath.Clean(root)

	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("--dir %s: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("--dir %s is not a directory", root)
	}
	return root, nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./cmd/tenet/ -v`
Expected: PASS for all six tests.

Then run: `gofmt -l cmd/ && go vet ./cmd/...`
Expected: no output from either command.

- [ ] **Step 5: Stdio smoke check**

Run:
```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_point_of_view","arguments":{}}}' \
  | go run ./cmd/tenet --dir testdata/project 2>/dev/null
```
Expected: stdout contains a JSON line with `"id":1` and a JSON line with `"id":2` whose text contains `Point of view`. Stdout has no non-JSON lines. If the protocol version string is rejected by the SDK, the implementer reports BLOCKED with the stdout it received.

---

### Task 4: CI workflow and README

**Model:** sonnet

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `README.md`

**Interfaces:**
- Consumes: the commands from Global Constraints and the `--dir` flag from Task 3.
- Produces: nothing consumed by later tasks.

- [ ] **Step 1: Write the CI workflow**

Create `.github/workflows/ci.yml` with exactly this content:
```yaml
name: CI

on:
  push:
  pull_request:

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: gofmt
        run: test -z "$(gofmt -l .)"
      - name: golangci-lint
        run: go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
      - name: test
        run: go test ./...
```

- [ ] **Step 2: Check the workflow parses**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml')); print('ok')"`
Expected: `ok`. If PyYAML is not installed, the implementer reports that the YAML was not machine-checked and does not install packages.

- [ ] **Step 3: Write the README**

Create `README.md` with these sections, in this order:

1. `# tenet`: one paragraph taken from the first paragraph of `GLOSSARY.md`, followed by a sentence saying the tool is at the walking-skeleton stage and has one tool, `get_point_of_view`.
2. `## Requirements`: the Go toolchain version in `go.mod`.
3. `## Run`: the commands `go run ./cmd/tenet --dir testdata/project` and `go build -o tenet ./cmd/tenet`, and a note that without `--dir` the working directory is the project root.
4. `## Connect Claude Code`: the command
   ```bash
   claude mcp add --transport stdio tenet -- /absolute/path/to/tenet --dir /absolute/path/to/your/project
   ```
   followed by the sentence "Then run `claude mcp list` and confirm `tenet` is connected." and a note that the project needs `design-system/point-of-view.md`, with a link to `testdata/project/design-system/point-of-view.md` as an example.
5. `## Develop`: the three checks, exactly:
   ```bash
   go test ./...
   test -z "$(gofmt -l .)"
   go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
   ```
6. `## Docs`: links to `GLOSSARY.md`, `docs/roadmap.md`, `docs/adr/`.

Do not describe features that are not built in this slice.

- [ ] **Step 4: Verify the README commands**

Run: `go run ./cmd/tenet --dir testdata/project </dev/null; echo "exit=$?"`
Expected: stderr shows `serving ... on stdio` and the process exits with `exit=0` once stdin closes.

Run: `go test ./... && test -z "$(gofmt -l .)" && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`
Expected: all three succeed with no lint findings.
