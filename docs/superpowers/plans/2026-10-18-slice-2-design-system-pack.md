# Slice 2: Design system pack format and validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.


> **Superseded in places.** The implemented code (commit ea7ee82) is authoritative. Where this plan's code blocks differ from it: the type `Pack` is named `DesignSystem`; `badByteIndex(off)` returns `max(off-1, 0)` with no data argument; `parseRules` excludes fence lines and fenced content from the rule statement; the BOM tests use the `\uFEFF` escape. The spec (also updated to `DesignSystem`) is the binding authority.

**Goal:** The server and a new `tenet validate` command read a `design-system/` folder (point of view, tokens JSON, rules Markdown, and component, template and flow Markdown items) through one loader that reports every problem with its file and line; `tenet init` scaffolds a valid empty folder.

**Architecture:** `internal/designsystem` gains a stdlib-only loader made of small per-file parsers (`tokens.go`, `rules.go`, `items.go`), one `Load(root)` that orchestrates them, and a `Scaffold(root)`. `cmd/tenet` dispatches subcommands (`validate`, `init`, and the unchanged default serve path) and prints problems as `path:line: message`. `internal/server` is unchanged.

**Tech Stack:** Go (toolchain from `go.mod`), `github.com/modelcontextprotocol/go-sdk` v1.8.0 (unchanged), `encoding/json` token streaming for line numbers, `regexp`, golangci-lint v2.15.0 run via `go run`.

**Spec:** [docs/superpowers/specs/2026-10-18-slice-2-design-system-pack-design.md](../specs/2026-10-18-slice-2-design-system-pack-design.md). Decisions D1–D28 there are binding; this plan refers to them by number.

## Global Constraints

- Module path: `github.com/tenet-ui/tenet`
- MCP SDK: `github.com/modelcontextprotocol/go-sdk` at exactly `v1.8.0`; do not change `go.mod` requirements
- `internal/designsystem` imports only the standard library; no MCP types
- stdout carries only MCP protocol during serving; all serving logs go to stderr. `validate` and `init` may write to stdout (they are not MCP sessions)
- Linter: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`
- Format: `test -z "$(gofmt -l .)"`
- Tests: `go test ./...`
- Point-of-view path (relative to root): `design-system/point-of-view.md`
- Tool name: `get_point_of_view`; server name `tenet`; `Version = "0.1.0"`
- Problem line numbers are 1-based; 0 means the whole file
- Problem text format: `path:line: message`, or `path: message` when line is 0
- Do not commit anywhere in this plan; leave changes in the working tree

## Review Focus

- A `tokens.json` saved with a UTF-8 byte-order mark (common from Windows editors) loads as normal, not as a JSON syntax error. Pinned by Task 2, `TestParseTokens` case "utf-8 BOM is ignored".
- `tokens.json` and `rules.md` with CRLF line endings load with the same tokens and rules and correct line numbers. Pinned by Task 2, case "CRLF line endings", and Task 3, case "CRLF line endings".
- Negative numbers and exponent notation (`-1`, `1e3`) are accepted as numeric token values, not rejected as malformed. Pinned by Task 2, case "negative and exponent numbers".
- A dangling symlink at `design-system` is treated as existing, so `tenet init` refuses and does not write through it. Pinned by Task 6, `TestScaffoldRefusesDanglingSymlink`.
- A component file named `Button.md` is reported, not silently skipped, so the author sees why it is missing from the pack. Pinned by Task 4, case "uppercase file name is reported".

---

### Task 1: Problem type, BOM and line helpers, and the point-of-view check in Load

**Model:** sonnet

**Files:**
- Create: `internal/designsystem/problem.go`
- Create: `internal/designsystem/problem_test.go`
- Create: `internal/designsystem/load.go`
- Create: `internal/designsystem/load_test.go`

`internal/designsystem/pointofview.go` (slice 1) is unchanged and is used by `Load` through `ReadPointOfView`.

**Interfaces:**
- Consumes: `ReadPointOfView`, `PointOfViewPath`, `ErrPointOfViewMissing`, `ErrPointOfViewEmpty` from `pointofview.go`.
- Produces:
  - `type Problem struct { File string; Line int; Message string }` and `func (p Problem) String() string`
  - `func lineAt(data []byte, off int64) int` (1-based line of byte offset `off`, clamped to `[0, len(data)]`)
  - `func stripBOM(data []byte) []byte`
  - `type Token struct { Category, Name, Value string; Line int }`
  - `type Rule struct { ID, Statement, Reason string; Line int }`
  - `type Item struct { ID, Title, Path string }`
  - `type Pack struct { PointOfView string; Tokens []Token; Rules []Rule; Components, Templates, Flows []Item }`
  - `const DirName = "design-system"`, `TokensPath = "design-system/tokens.json"`, `RulesPath = "design-system/rules.md"`, `ComponentsDir = "design-system/components"`, `TemplatesDir = "design-system/templates"`, `FlowsDir = "design-system/flows"`
  - `func Load(root string) (Pack, []Problem)`: this task implements the directory and point-of-view checks; Task 5 extends it.
  - Test helpers used by later tasks, defined in `load_test.go`:
    - `func writeTree(t *testing.T, root string, files map[string]string)`
    - `type wantProblem struct { Line int; Contains string }`
    - `func checkProblems(t *testing.T, got []Problem, want []wantProblem)`: exact count, in order; line must match; message must contain `Contains`
    - `func problemsFor(ps []Problem, file string) []Problem`

- [ ] **Step 1: Write the failing tests for the helpers**

Create `internal/designsystem/problem_test.go`:

```go
package designsystem

import (
	"testing"
)

func TestProblemString(t *testing.T) {
	if got := (Problem{File: "a.json", Line: 3, Message: "bad"}).String(); got != "a.json:3: bad" {
		t.Errorf("with line: got %q", got)
	}
	if got := (Problem{File: "a.json", Message: "missing"}).String(); got != "a.json: missing" {
		t.Errorf("without line: got %q", got)
	}
}

func TestLineAt(t *testing.T) {
	data := []byte("ab\ncd\nef")
	tests := []struct {
		off  int64
		want int
	}{
		{0, 1},
		{2, 1},
		{3, 2},
		{6, 3},
		{100, 3},
		{-1, 1},
	}
	for _, tt := range tests {
		if got := lineAt(data, tt.off); got != tt.want {
			t.Errorf("lineAt(%d) = %d, want %d", tt.off, got, tt.want)
		}
	}
}

func TestStripBOM(t *testing.T) {
	if got := string(stripBOM([]byte("﻿x"))); got != "x" {
		t.Errorf("with BOM: got %q", got)
	}
	if got := string(stripBOM([]byte("x"))); got != "x" {
		t.Errorf("without BOM: got %q", got)
	}
}
```

Create `internal/designsystem/load_test.go` with the shared helpers and the point-of-view tests:

```go
package designsystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree creates each file under root, making parent directories as needed.
func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// wantProblem describes one expected problem: its line (0 for a file-level problem) and a substring of its message.
type wantProblem struct {
	Line     int
	Contains string
}

// checkProblems fails the test unless got holds exactly the wanted problems, in order.
func checkProblems(t *testing.T, got []Problem, want []wantProblem) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d problems %v, want %d", len(got), got, len(want))
	}
	for i, w := range want {
		if got[i].Line != w.Line || !strings.Contains(got[i].Message, w.Contains) {
			t.Errorf("problem %d = %v, want line %d with message containing %q", i, got[i], w.Line, w.Contains)
		}
	}
}

// problemsFor returns the problems whose File is file.
func problemsFor(ps []Problem, file string) []Problem {
	var out []Problem
	for _, p := range ps {
		if p.File == file {
			out = append(out, p)
		}
	}
	return out
}

func TestLoadDesignSystemMissing(t *testing.T) {
	_, problems := Load(t.TempDir())
	checkProblems(t, problems, []wantProblem{{0, "directory not found"}})
}

func TestLoadDesignSystemIsFile(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{"design-system": "x"})
	_, problems := Load(root)
	checkProblems(t, problems, []wantProblem{{0, "not a directory"}})
}

func TestLoadPointOfView(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantPOV bool
		want    []wantProblem
	}{
		{
			name:    "valid",
			files:   map[string]string{"design-system/point-of-view.md": "# Brand\n"},
			wantPOV: true,
		},
		{
			name:  "file missing",
			files: map[string]string{"design-system/keep.txt": ""},
			want:  []wantProblem{{0, "file not found"}},
		},
		{
			name:  "whitespace only",
			files: map[string]string{"design-system/point-of-view.md": " \n\t\n"},
			want:  []wantProblem{{1, "file is empty"}},
		},
		{
			name:  "point of view is a directory",
			files: map[string]string{"design-system/point-of-view.md/keep.txt": "x"},
			want:  []wantProblem{{0, "could not read"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeTree(t, root, tt.files)
			pack, problems := Load(root)
			checkProblems(t, problemsFor(problems, PointOfViewPath), tt.want)
			if (pack.PointOfView != "") != tt.wantPOV {
				t.Errorf("PointOfView = %q, want non-empty=%v", pack.PointOfView, tt.wantPOV)
			}
		})
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/designsystem/ -run 'TestProblemString|TestLineAt|TestStripBOM|TestLoad' -v`
Expected: FAIL with build errors such as `undefined: Problem`, `undefined: Load`.

- [ ] **Step 3: Write `problem.go`**

Create `internal/designsystem/problem.go`:

```go
package designsystem

import (
	"bytes"
	"fmt"
)

// Problem is one validation finding. Line is 1-based; 0 means the whole file.
type Problem struct {
	File    string
	Line    int
	Message string
}

// String formats the problem as "file:line: message", or "file: message" when Line is 0.
func (p Problem) String() string {
	if p.Line == 0 {
		return fmt.Sprintf("%s: %s", p.File, p.Message)
	}
	return fmt.Sprintf("%s:%d: %s", p.File, p.Line, p.Message)
}

// lineAt returns the 1-based line of byte offset off in data. Offsets outside data are clamped.
func lineAt(data []byte, off int64) int {
	if off < 0 {
		off = 0
	}
	if off > int64(len(data)) {
		off = int64(len(data))
	}
	return 1 + bytes.Count(data[:off], []byte("\n"))
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// stripBOM removes a leading UTF-8 byte-order mark. Line numbers are unaffected.
func stripBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, utf8BOM)
}
```

- [ ] **Step 4: Write `load.go` with the types and the directory and point-of-view checks**

Create `internal/designsystem/load.go`:

```go
package designsystem

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	// DirName is the design system folder, relative to the project root.
	DirName = "design-system"
	// TokensPath is the tokens file, relative to the project root.
	TokensPath = "design-system/tokens.json"
	// RulesPath is the rules file, relative to the project root.
	RulesPath = "design-system/rules.md"
	// ComponentsDir, TemplatesDir and FlowsDir each hold one Markdown file per item.
	ComponentsDir = "design-system/components"
	TemplatesDir  = "design-system/templates"
	FlowsDir      = "design-system/flows"
)

// Token is one entry of tokens.json.
type Token struct {
	Category string
	Name     string
	Value    string // string value, or the number's literal text
	Line     int
}

// Rule is one "## <id>" entry of rules.md.
type Rule struct {
	ID        string
	Statement string // body lines other than the Reason: line, joined with a space
	Reason    string
	Line      int // line of the "## <id>" heading
}

// Item is one Markdown file in components/, templates/ or flows/.
type Item struct {
	ID    string // file name without ".md"
	Title string // text of the first "# " heading
	Path  string // slash-separated, relative to the project root
}

// Pack is everything the loader read from design-system/.
type Pack struct {
	PointOfView string // empty when the file has problems
	Tokens      []Token
	Rules       []Rule
	Components  []Item
	Templates   []Item
	Flows       []Item
}

// Load reads root/design-system and returns the pack it could read and every
// problem it found. A non-empty problems slice means the pack is not valid,
// but Load still returns whatever it could parse.
func Load(root string) (Pack, []Problem) {
	var pack Pack
	info, err := os.Stat(filepath.Join(root, DirName))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return pack, []Problem{{File: DirName, Message: "directory not found; run tenet init"}}
	case err != nil:
		return pack, []Problem{{File: DirName, Message: "could not read: " + err.Error()}}
	case !info.IsDir():
		return pack, []Problem{{File: DirName, Message: "not a directory"}}
	}

	var problems []Problem
	text, err := ReadPointOfView(root)
	switch {
	case errors.Is(err, ErrPointOfViewMissing):
		problems = append(problems, Problem{File: PointOfViewPath, Message: "file not found"})
	case errors.Is(err, ErrPointOfViewEmpty):
		problems = append(problems, Problem{File: PointOfViewPath, Line: 1, Message: "file is empty"})
	case err != nil:
		problems = append(problems, Problem{File: PointOfViewPath, Message: "could not read: " + err.Error()})
	default:
		pack.PointOfView = text
	}
	return pack, problems
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/designsystem/ -run 'TestProblemString|TestLineAt|TestStripBOM|TestLoad' -v`
Expected: PASS for all listed tests.

- [ ] **Step 6: Verify the package builds and vets**

Run: `go vet ./internal/designsystem/ && go test ./...`
Expected: no output from vet; all packages `ok` or `no test files`.

---

### Task 2: Tokens parser (`tokens.json`)

**Model:** sonnet

**Files:**
- Create: `internal/designsystem/tokens.go`
- Create: `internal/designsystem/tokens_test.go`

**Interfaces:**
- Consumes: `Token`, `Problem`, `lineAt`, `stripBOM` from Task 1.
- Produces: `func parseTokens(file string, data []byte) ([]Token, []Problem)`. `file` is the project-relative path used in problems (the caller passes `TokensPath`). Tokens are returned in file order. Problems are returned in file order.

Behaviour (D4–D8, D23 and the error table of the spec):
- Empty or whitespace-only input → one problem at line 1, `file is empty; write {} for no tokens`.
- Top level must be `{`. Otherwise one problem at the first token's line, `top level must be an object of categories`.
- Category names and token names match `^[a-z][a-z0-9-]*$`.
- A category value must be an object. Otherwise problem at the key's line, `category "<cat>" must be an object of tokens`. The value is skipped.
- A token value must be a JSON string or number. Bool, null, array, and object are problems at the key's line: `token "<cat>.<name>" must be a string or number`. The value is skipped.
- An empty string value is a problem at the key's line: `token "<cat>.<name>" is empty`.
- Duplicate category: problem at the second key's line, `duplicate category "<cat>"`. Its tokens are not kept, but its members are still checked.
- Duplicate token within a category: problem at the second key's line, `duplicate token "<cat>.<name>"`. The first is kept.
- Invalid category name: problem at the key's line, `category name "<cat>" must be lowercase letters, digits and dashes`. Its tokens are checked but not kept.
- Invalid token name: problem at the key's line, `token name "<name>" must be lowercase letters, digits and dashes`.
- Trailing content after the top-level object: problem at the first non-space byte's line, `unexpected content after the top-level object`.
- Decoding error: one problem, `invalid JSON: <decoder message>`, at the line of the bad byte. Parsing stops. Earlier tokens are kept.
- Numbers keep their literal text (`json.Number`).

- [ ] **Step 1: Write the failing test**

Create `internal/designsystem/tokens_test.go`:

```go
package designsystem

import (
	"fmt"
	"strings"
	"testing"
)

// formatTokens renders each token as "category.name=value@line" for comparison.
func formatTokens(ts []Token) []string {
	out := make([]string, 0, len(ts))
	for _, tk := range ts {
		out = append(out, fmt.Sprintf("%s.%s=%s@%d", tk.Category, tk.Name, tk.Value, tk.Line))
	}
	return out
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}

func TestParseTokens(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantTokens   []string
		wantProblems []wantProblem
	}{
		{
			name:       "valid two categories",
			data:       `{"color":{"ink":"#1d2433"},"space":{"unit":4,"half":1.5}}`,
			wantTokens: []string{"color.ink=#1d2433@1", "space.unit=4@1", "space.half=1.5@1"},
		},
		{
			name:       "negative and exponent numbers",
			data:       `{"space":{"neg":-1,"big":1e3}}`,
			wantTokens: []string{"space.neg=-1@1", "space.big=1e3@1"},
		},
		{
			name:         "empty file",
			data:         "",
			wantProblems: []wantProblem{{1, "file is empty"}},
		},
		{
			name:         "whitespace only",
			data:         "  \n \n",
			wantProblems: []wantProblem{{1, "file is empty"}},
		},
		{
			name:       "empty object is valid",
			data:       "{}\n",
			wantTokens: nil,
		},
		{
			name:       "utf-8 BOM is ignored",
			data:       "﻿{\"color\":{\"ink\":\"#fff\"}}",
			wantTokens: []string{"color.ink=#fff@1"},
		},
		{
			name:       "CRLF line endings",
			data:       "{\r\n  \"color\": {\r\n    \"ink\": \"#fff\"\r\n  }\r\n}\r\n",
			wantTokens: []string{"color.ink=#fff@3"},
		},
		{
			name:         "top level is an array",
			data:         `[]`,
			wantProblems: []wantProblem{{1, "top level must be an object"}},
		},
		{
			name:         "category is not an object",
			data:         `{"color": "red"}`,
			wantProblems: []wantProblem{{1, `category "color" must be an object`}},
		},
		{
			name:         "nested object value",
			data:         `{"color": {"brand": {"primary": "#fff"}}}`,
			wantProblems: []wantProblem{{1, `token "color.brand" must be a string or number`}},
		},
		{
			name:         "boolean value",
			data:         `{"color":{"on":true}}`,
			wantProblems: []wantProblem{{1, `token "color.on" must be a string or number`}},
		},
		{
			name:         "null value",
			data:         `{"color":{"on":null}}`,
			wantProblems: []wantProblem{{1, `token "color.on" must be a string or number`}},
		},
		{
			name:         "array value",
			data:         `{"color":{"list":[1,2]}}`,
			wantProblems: []wantProblem{{1, `token "color.list" must be a string or number`}},
		},
		{
			name:         "empty string value",
			data:         `{"color":{"ink":""}}`,
			wantProblems: []wantProblem{{1, `token "color.ink" is empty`}},
		},
		{
			name:         "invalid token name",
			data:         `{"color":{"Ink":"#fff"}}`,
			wantProblems: []wantProblem{{1, `token name "Ink" must be lowercase`}},
		},
		{
			name:         "invalid category name",
			data:         `{"Color":{"ink":"#fff"}}`,
			wantProblems: []wantProblem{{1, `category name "Color" must be lowercase`}},
		},
		{
			name: "duplicate category reports second key and keeps first",
			data: joinLines(
				"{",
				`  "color": {"ink": "#fff"},`,
				`  "color": {"paper": "#000"}`,
				"}",
			),
			wantTokens:   []string{"color.ink=#fff@2"},
			wantProblems: []wantProblem{{3, `duplicate category "color"`}},
		},
		{
			name: "duplicate token reports second key and keeps first",
			data: joinLines(
				"{",
				`  "color": {`,
				`    "ink": "#fff",`,
				`    "ink": "#000"`,
				"  }",
				"}",
			),
			wantTokens:   []string{"color.ink=#fff@3"},
			wantProblems: []wantProblem{{4, `duplicate token "color.ink"`}},
		},
		{
			name:         "trailing content after the object",
			data:         "{}\n{}",
			wantProblems: []wantProblem{{2, "unexpected content after the top-level object"}},
		},
		{
			name:         "truncated file",
			data:         `{"color": {"ink": "#fff"`,
			wantTokens:   []string{"color.ink=#fff@1"},
			wantProblems: []wantProblem{{1, "invalid JSON"}},
		},
		{
			name: "syntax error reports the bad byte's line",
			data: joinLines(
				"{",
				`  "color": {`,
				`    "ink" "#fff"`,
				"  }",
				"}",
			),
			wantProblems: []wantProblem{{3, "invalid JSON"}},
		},
		{
			name: "syntax error at the start of a line",
			data: joinLines(
				"{",
				`  "color": {`,
				`    "ink"`,
				`"#fff"`,
				"  }",
				"}",
			),
			wantProblems: []wantProblem{{4, "invalid JSON"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, problems := parseTokens(TokensPath, []byte(tt.data))
			if diff := strings.Join(formatTokens(got), "|"); diff != strings.Join(tt.wantTokens, "|") {
				t.Errorf("tokens = %v, want %v", formatTokens(got), tt.wantTokens)
			}
			for _, p := range problems {
				if p.File != TokensPath {
					t.Errorf("problem file = %q, want %q", p.File, TokensPath)
				}
			}
			checkProblems(t, problems, tt.wantProblems)
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/designsystem/ -run TestParseTokens -v`
Expected: FAIL with `undefined: parseTokens`.

- [ ] **Step 3: Write `tokens.go`**

Create `internal/designsystem/tokens.go`:

```go
package designsystem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"unicode"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// tokenParser walks tokens.json with a streaming decoder so every problem can
// carry the line of the token it describes.
type tokenParser struct {
	file     string
	data     []byte
	dec      *json.Decoder
	tokens   []Token
	problems []Problem
}

// parseTokens reads a tokens.json file. file is the project-relative path used in problems.
func parseTokens(file string, data []byte) ([]Token, []Problem) {
	data = stripBOM(data)
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, []Problem{{File: file, Line: 1, Message: "file is empty; write {} for no tokens"}}
	}
	p := &tokenParser{file: file, data: data, dec: json.NewDecoder(bytes.NewReader(data))}
	p.dec.UseNumber()
	p.parseDocument()
	return p.tokens, p.problems
}

func (p *tokenParser) add(line int, format string, args ...any) {
	p.problems = append(p.problems, Problem{File: p.file, Line: line, Message: fmt.Sprintf(format, args...)})
}

// line is the line of the decoder's current position, which is just after the last token read.
func (p *tokenParser) line() int {
	return lineAt(p.data, p.dec.InputOffset())
}

// next reads the next token. A decoding error becomes one problem and ok is false.
func (p *tokenParser) next() (json.Token, bool) {
	tok, err := p.dec.Token()
	if err == nil {
		return tok, true
	}
	line := lineAt(p.data, int64(len(p.data)))
	var se *json.SyntaxError
	if errors.As(err, &se) {
		line = lineAt(p.data, badByteIndex(p.data, se.Offset))
	}
	p.add(line, "invalid JSON: %s", err)
	return nil, false
}

// badByteIndex returns the index of the byte a syntax error points at. The
// decoder is inconsistent: sometimes Offset is one past the bad byte, sometimes
// it is the bad byte itself. A byte before Offset that is not white space means
// Offset is one past it.
func badByteIndex(data []byte, off int64) int64 {
	i := off - 1
	if i < 0 || i >= int64(len(data)) || unicode.IsSpace(rune(data[i])) {
		return off
	}
	return i
}

func (p *tokenParser) parseDocument() {
	tok, ok := p.next()
	if !ok {
		return
	}
	if d, isDelim := tok.(json.Delim); !isDelim || d != '{' {
		p.add(p.line(), "top level must be an object of categories")
		return
	}
	seen := map[string]bool{}
	for p.dec.More() {
		key, ok := p.next()
		if !ok {
			return
		}
		cat, _ := key.(string)
		if !p.parseCategory(cat, p.line(), seen) {
			return
		}
	}
	if tok, ok := p.next(); !ok || tok != json.Delim('}') {
		return
	}
	if _, err := p.dec.Token(); err != io.EOF {
		rest := p.data[p.dec.InputOffset():]
		trimmed := bytes.TrimLeft(rest, " \t\r\n")
		offset := p.dec.InputOffset() + int64(len(rest)-len(trimmed))
		p.add(lineAt(p.data, offset), "unexpected content after the top-level object")
	}
}

// parseCategory reads the value of one category key. Tokens are kept only when
// the category name is valid and not a duplicate; members are always checked.
// It returns false when decoding stopped.
func (p *tokenParser) parseCategory(cat string, keyLine int, seen map[string]bool) bool {
	dup := seen[cat]
	seen[cat] = true
	valid := namePattern.MatchString(cat)
	if !valid {
		p.add(keyLine, "category name %q must be lowercase letters, digits and dashes", cat)
	}
	if dup {
		p.add(keyLine, "duplicate category %q", cat)
	}
	keep := valid && !dup

	tok, ok := p.next()
	if !ok {
		return false
	}
	if d, isDelim := tok.(json.Delim); !isDelim || d != '{' {
		p.add(keyLine, "category %q must be an object of tokens", cat)
		return p.skipValue(tok)
	}
	names := map[string]bool{}
	for p.dec.More() {
		key, ok := p.next()
		if !ok {
			return false
		}
		name, _ := key.(string)
		if !p.parseToken(cat, name, p.line(), keep, names) {
			return false
		}
	}
	_, ok = p.next() // the closing brace of the category object
	return ok
}

// parseToken reads one token value and records it when keep is true and it is valid.
func (p *tokenParser) parseToken(cat, name string, line int, keep bool, names map[string]bool) bool {
	qualified := cat + "." + name
	vtok, ok := p.next()
	if !ok {
		return false
	}
	value, isScalar := scalarText(vtok)
	if !p.skipValue(vtok) {
		return false
	}
	if names[name] {
		p.add(line, "duplicate token %q", qualified)
		return true
	}
	names[name] = true
	switch {
	case !namePattern.MatchString(name):
		p.add(line, "token name %q must be lowercase letters, digits and dashes", name)
	case !isScalar:
		p.add(line, "token %q must be a string or number", qualified)
	case value == "":
		p.add(line, "token %q is empty", qualified)
	case keep:
		p.tokens = append(p.tokens, Token{Category: cat, Name: name, Value: value, Line: line})
	}
	return true
}

// skipValue consumes the rest of a value whose first token has already been read.
// Scalars need nothing more. Objects and arrays are read through their closing bracket.
func (p *tokenParser) skipValue(first json.Token) bool {
	d, isDelim := first.(json.Delim)
	if !isDelim || (d != '{' && d != '[') {
		return true
	}
	depth := 1
	for depth > 0 {
		tok, ok := p.next()
		if !ok {
			return false
		}
		if d, isDelim := tok.(json.Delim); isDelim {
			switch d {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		}
	}
	return true
}

// scalarText returns the text of a string or number token.
func scalarText(tok json.Token) (string, bool) {
	switch v := tok.(type) {
	case string:
		return v, true
	case json.Number:
		return v.String(), true
	}
	return "", false
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/designsystem/ -run TestParseTokens -v`
Expected: PASS for every case. If a syntax-error line case fails, check the decoder's offset with a one-off print before changing `badByteIndex`; the expected lines in the table are the spec's required behaviour (D23).

- [ ] **Step 5: Run the package tests**

Run: `go test ./internal/designsystem/`
Expected: `ok`.

---

### Task 3: Rules parser (`rules.md`)

**Model:** sonnet

**Files:**
- Create: `internal/designsystem/rules.go`
- Create: `internal/designsystem/rules_test.go`

**Interfaces:**
- Consumes: `Rule`, `Problem`, `stripBOM` from Task 1.
- Produces: `func parseRules(file string, data []byte) ([]Rule, []Problem)`. Rules are returned in heading order. Problems are returned in line order within each rule.

Behaviour (D9–D12 and the spec's error table):
- A line equal to `##`, or starting with `## `, outside a fenced code block starts a rule. Its id is the text after `##`, trimmed. A fence is any line whose trimmed text starts with three backticks; it toggles fence state.
- Text before the first rule is ignored. Zero rules is valid and produces no problem.
- Id must match `^[a-z0-9]+([.-][a-z0-9]+)*$`. Otherwise problem at the heading line, `rule id "<id>" must be lowercase letters and digits joined by dots or dashes`, and the rule is dropped.
- A duplicate id: problem at the second heading line, `duplicate rule id "<id>" (first at line N)`, and the second rule is dropped.
- Otherwise the rule is kept. Body lines whose trimmed text starts with `Reason:` are reason lines; others with non-empty trimmed text form the statement, joined with a single space.
- No statement: problem at the heading line, `rule "<id>" has no statement`.
- No reason line: problem at the heading line, `rule "<id>" has no Reason: line`. Reason is empty.
- Reason line with empty text: problem at that line, `rule "<id>" has an empty Reason: line`.
- A second reason line: problem at that line, `rule "<id>" has more than one Reason: line`. The first is used.
- CRLF is handled by trimming `\r` from each line. A BOM is stripped first.

- [ ] **Step 1: Write the failing test**

Create `internal/designsystem/rules_test.go`:

```go
package designsystem

import (
	"fmt"
	"strings"
	"testing"
)

// formatRules renders each rule as "id@line|statement|reason" for comparison.
func formatRules(rs []Rule) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, fmt.Sprintf("%s@%d|%s|%s", r.ID, r.Line, r.Statement, r.Reason))
	}
	return out
}

func TestParseRules(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantRules    []string
		wantProblems []wantProblem
	}{
		{
			name:      "zero entries is valid",
			data:      "# Rules\n\nIntro text.\n",
			wantRules: nil,
		},
		{
			name:      "empty file is valid",
			data:      "",
			wantRules: nil,
		},
		{
			name: "two valid entries",
			data: joinLines(
				"# Rules",
				"",
				"## color.tokens-only",
				"Components take colours from tokens.",
				"Reason: Raw values drift.",
				"",
				"## spacing.unit-multiples",
				"Spacing is a multiple of the unit.",
				"Reason: Off-grid spacing looks unplanned.",
			),
			wantRules: []string{
				"color.tokens-only@3|Components take colours from tokens.|Raw values drift.",
				"spacing.unit-multiples@7|Spacing is a multiple of the unit.|Off-grid spacing looks unplanned.",
			},
		},
		{
			name: "multi-line statement is joined with a space",
			data: joinLines(
				"## a.b",
				"First line.",
				"Second line.",
				"Reason: because.",
			),
			wantRules: []string{"a.b@1|First line. Second line.|because."},
		},
		{
			name: "CRLF line endings",
			data: "## a\r\nS.\r\nReason: r.\r\n",
			wantRules: []string{"a@1|S.|r."},
		},
		{
			name:      "UTF-8 BOM is ignored",
			data:      "﻿## a\nS.\nReason: r\n",
			wantRules: []string{"a@1|S.|r"},
		},
		{
			name: "duplicate id reports second heading and drops it",
			data: joinLines(
				"## a",
				"S.",
				"Reason: r",
				"## a",
				"S.",
				"Reason: r",
			),
			wantRules:    []string{"a@1|S.|r"},
			wantProblems: []wantProblem{{4, `duplicate rule id "a" (first at line 1)`}},
		},
		{
			name: "invalid ids are reported and dropped",
			data: joinLines(
				"## Color",
				"S.",
				"Reason: r",
				"## a..b",
				"S.",
				"Reason: r",
				"## -x",
				"S.",
				"Reason: r",
				"## a b",
				"S.",
				"Reason: r",
				"## ",
				"S.",
				"Reason: r",
			),
			wantProblems: []wantProblem{
				{1, `rule id "Color" must be`},
				{4, `rule id "a..b" must be`},
				{7, `rule id "-x" must be`},
				{10, `rule id "a b" must be`},
				{13, `rule id "" must be`},
			},
		},
		{
			name:         "no statement",
			data:         "## a\nReason: x\n",
			wantRules:    []string{"a@1||x"},
			wantProblems: []wantProblem{{1, `rule "a" has no statement`}},
		},
		{
			name:         "no Reason line",
			data:         "## a\nStatement.\n",
			wantRules:    []string{"a@1|Statement.|"},
			wantProblems: []wantProblem{{1, `rule "a" has no Reason: line`}},
		},
		{
			name:         "empty Reason line",
			data:         "## a\nStatement.\nReason:   \n",
			wantRules:    []string{"a@1|Statement.|"},
			wantProblems: []wantProblem{{3, `rule "a" has an empty Reason: line`}},
		},
		{
			name:         "repeated Reason line keeps the first",
			data:         "## a\nS.\nReason: x\nReason: y\n",
			wantRules:    []string{"a@1|S.|x"},
			wantProblems: []wantProblem{{4, `rule "a" has more than one Reason: line`}},
		},
		{
			name: "heading inside a fenced code block is not a rule",
			data: joinLines(
				"## a",
				"S.",
				"Reason: r.",
				"",
				"```md",
				"## not-a-rule",
				"```",
			),
			wantRules: []string{"a@1|S.|r."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, problems := parseRules(RulesPath, []byte(tt.data))
			if g, w := strings.Join(formatRules(got), "\n"), strings.Join(tt.wantRules, "\n"); g != w {
				t.Errorf("rules =\n%s\nwant\n%s", g, w)
			}
			for _, p := range problems {
				if p.File != RulesPath {
					t.Errorf("problem file = %q, want %q", p.File, RulesPath)
				}
			}
			checkProblems(t, problems, tt.wantProblems)
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/designsystem/ -run TestParseRules -v`
Expected: FAIL with `undefined: parseRules`.

- [ ] **Step 3: Write `rules.go`**

Create `internal/designsystem/rules.go`:

```go
package designsystem

import (
	"fmt"
	"regexp"
	"strings"
)

var ruleIDPattern = regexp.MustCompile(`^[a-z0-9]+([.-][a-z0-9]+)*$`)

// numberedLine is one body line of a rule entry with its 1-based line number.
type numberedLine struct {
	no   int
	text string
}

// ruleEntry is one "## <id>" section before its body is interpreted.
type ruleEntry struct {
	id   string
	line int
	body []numberedLine
}

// parseRules reads rules.md. A line "## <id>" outside a fenced code block starts
// a rule. Its body holds a statement and one "Reason:" line.
func parseRules(file string, data []byte) ([]Rule, []Problem) {
	var (
		entries []*ruleEntry
		current *ruleEntry
		inFence bool
	)
	for i, raw := range strings.Split(string(stripBOM(data)), "\n") {
		text := strings.TrimRight(raw, "\r")
		if strings.HasPrefix(strings.TrimSpace(text), "```") {
			inFence = !inFence
		}
		if !inFence && (text == "##" || strings.HasPrefix(text, "## ")) {
			current = &ruleEntry{id: strings.TrimSpace(strings.TrimPrefix(text, "##")), line: i + 1}
			entries = append(entries, current)
			continue
		}
		if current != nil {
			current.body = append(current.body, numberedLine{no: i + 1, text: text})
		}
	}

	var (
		rules    []Rule
		problems []Problem
		seen     = map[string]int{}
	)
	add := func(line int, format string, args ...any) {
		problems = append(problems, Problem{File: file, Line: line, Message: fmt.Sprintf(format, args...)})
	}
	for _, e := range entries {
		if !ruleIDPattern.MatchString(e.id) {
			add(e.line, "rule id %q must be lowercase letters and digits joined by dots or dashes", e.id)
			continue
		}
		if first, dup := seen[e.id]; dup {
			add(e.line, "duplicate rule id %q (first at line %d)", e.id, first)
			continue
		}
		seen[e.id] = e.line

		var (
			statement  []string
			reason     string
			reasonLine int
			reasons    int
		)
		for _, l := range e.body {
			t := strings.TrimSpace(l.text)
			if strings.HasPrefix(t, "Reason:") {
				reasons++
				if reasons == 1 {
					reason = strings.TrimSpace(strings.TrimPrefix(t, "Reason:"))
					reasonLine = l.no
				} else {
					add(l.no, "rule %q has more than one Reason: line", e.id)
				}
				continue
			}
			if t != "" {
				statement = append(statement, t)
			}
		}
		if len(statement) == 0 {
			add(e.line, "rule %q has no statement", e.id)
		}
		if reasons == 0 {
			add(e.line, "rule %q has no Reason: line", e.id)
		} else if reason == "" {
			add(reasonLine, "rule %q has an empty Reason: line", e.id)
		}
		rules = append(rules, Rule{
			ID:        e.id,
			Statement: strings.Join(statement, " "),
			Reason:    reason,
			Line:      e.line,
		})
	}
	return rules, problems
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/designsystem/ -run TestParseRules -v`
Expected: PASS for every case.

- [ ] **Step 5: Run the package tests**

Run: `go test ./internal/designsystem/`
Expected: `ok`.

---

### Task 4: Item loader for components, templates and flows

**Model:** haiku

**Files:**
- Create: `internal/designsystem/items.go`
- Create: `internal/designsystem/items_test.go`

**Interfaces:**
- Consumes: `Item`, `Problem`, `stripBOM`, `DirName`, `ComponentsDir`, `TemplatesDir`, `FlowsDir`, `writeTree` and `checkProblems` from Tasks 1 and the helpers in `load_test.go`.
- Produces: `func loadItems(root, relDir string) ([]Item, []Problem)`. `relDir` is slash-separated and relative to root (`ComponentsDir`, `TemplatesDir`, or `FlowsDir`). Items are in file-name order, which is the order `os.ReadDir` returns.

Behaviour (D13–D16):
- Missing directory: no items, no problems.
- Path exists but is not a directory: one problem, File `relDir`, line 0, `not a directory`.
- Entries that are directories, or whose name does not end in `.md`, are ignored.
- Name (without `.md`) must match `^[a-z0-9]+(-[a-z0-9]+)*$`. Otherwise problem, File `relDir + "/" + name`, line 0, `file name must be lowercase kebab-case, for example "primary-button.md"`, and the item is left out.
- Content: after stripping a BOM, if the trimmed content is empty, problem at line 1, `file is empty; start with a "# <title>" line`.
- Otherwise the first non-blank line must be `# ` followed by a non-empty title. Otherwise problem at that line, `first line must be a "# <title>" heading`.
- The title is the trimmed text after `# `.

- [ ] **Step 1: Write the failing test**

Create `internal/designsystem/items_test.go`:

```go
package designsystem

import (
	"fmt"
	"testing"
)

// formatItems renders each item as "id|title|path" for comparison.
func formatItems(items []Item) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, fmt.Sprintf("%s|%s|%s", it.ID, it.Title, it.Path))
	}
	return out
}

func TestLoadItems(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		wantItems    []string
		wantProblems []wantProblem
	}{
		{
			name: "two items sorted by file name",
			files: map[string]string{
				"design-system/components/card.md":   "# Card\n\nA surface.\n",
				"design-system/components/button.md": "# Button\n\nAn action.\n",
			},
			wantItems: []string{
				"button|Button|design-system/components/button.md",
				"card|Card|design-system/components/card.md",
			},
		},
		{
			name:      "missing folder has no items and no problems",
			files:     map[string]string{"design-system/keep.txt": ""},
			wantItems: nil,
		},
		{
			name:         "folder is a file",
			files:        map[string]string{"design-system/components": "x"},
			wantProblems: []wantProblem{{0, "not a directory"}},
		},
		{
			name: "non-md files and subdirectories are ignored",
			files: map[string]string{
				"design-system/components/.gitkeep":        "",
				"design-system/components/notes.txt":       "x",
				"design-system/components/nested/x.md":     "# Nested\n",
				"design-system/components/button.md":       "# Button\n",
			},
			wantItems: []string{"button|Button|design-system/components/button.md"},
		},
		{
			name: "uppercase file name is reported",
			files: map[string]string{
				"design-system/components/Button.md": "# Button\n",
			},
			wantProblems: []wantProblem{{0, "file name must be lowercase kebab-case"}},
		},
		{
			name: "empty file",
			files: map[string]string{
				"design-system/components/button.md": " \n\n",
			},
			wantProblems: []wantProblem{{1, "file is empty"}},
		},
		{
			name: "first non-blank line is not a heading",
			files: map[string]string{
				"design-system/components/button.md": "\n\nButton\n# Later\n",
			},
			wantProblems: []wantProblem{{3, `first line must be a "# <title>" heading`}},
		},
		{
			name: "heading with empty title",
			files: map[string]string{
				"design-system/components/button.md": "#  \n",
			},
			wantProblems: []wantProblem{{1, `first line must be a "# <title>" heading`}},
		},
		{
			name: "UTF-8 BOM before the heading is ignored",
			files: map[string]string{
				"design-system/components/button.md": "﻿# Button\n",
			},
			wantItems: []string{"button|Button|design-system/components/button.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeTree(t, root, tt.files)
			items, problems := loadItems(root, ComponentsDir)
			if g, w := fmt.Sprint(formatItems(items)), fmt.Sprint(tt.wantItems); g != w {
				t.Errorf("items = %s, want %s", g, w)
			}
			checkProblems(t, problems, tt.wantProblems)
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/designsystem/ -run TestLoadItems -v`
Expected: FAIL with `undefined: loadItems`.

- [ ] **Step 3: Write `items.go`**

Create `internal/designsystem/items.go`:

```go
package designsystem

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var itemIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// loadItems reads the Markdown items directly inside root/relDir. relDir is
// slash-separated and relative to root. A missing directory yields no items.
func loadItems(root, relDir string) ([]Item, []Problem) {
	abs := filepath.Join(root, filepath.FromSlash(relDir))
	info, err := os.Stat(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, []Problem{{File: relDir, Message: "could not read: " + err.Error()}}
	}
	if !info.IsDir() {
		return nil, []Problem{{File: relDir, Message: "not a directory"}}
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, []Problem{{File: relDir, Message: "could not read: " + err.Error()}}
	}

	var (
		items    []Item
		problems []Problem
	)
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".md" {
			continue
		}
		file := relDir + "/" + e.Name()
		id := strings.TrimSuffix(e.Name(), ".md")
		if !itemIDPattern.MatchString(id) {
			problems = append(problems, Problem{File: file, Message: `file name must be lowercase kebab-case, for example "primary-button.md"`})
			continue
		}
		data, err := os.ReadFile(filepath.Join(abs, e.Name()))
		if err != nil {
			problems = append(problems, Problem{File: file, Message: "could not read: " + err.Error()})
			continue
		}
		title, line, msg := itemTitle(stripBOM(data))
		if msg != "" {
			problems = append(problems, Problem{File: file, Line: line, Message: msg})
			continue
		}
		items = append(items, Item{ID: id, Title: title, Path: file})
	}
	return items, problems
}

// itemTitle returns the title from the first non-blank line, which must be a
// level-1 heading. On failure it returns the line to report and a message.
func itemTitle(data []byte) (title string, line int, problem string) {
	if strings.TrimSpace(string(data)) == "" {
		return "", 1, `file is empty; start with a "# <title>" line`
	}
	for i, raw := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "# ") && strings.TrimSpace(t[2:]) != "" {
			return strings.TrimSpace(t[2:]), 0, ""
		}
		return "", i + 1, `first line must be a "# <title>" heading`
	}
	return "", 1, `file is empty; start with a "# <title>" line`
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/designsystem/ -run TestLoadItems -v`
Expected: PASS for every case.

- [ ] **Step 5: Run the package tests**

Run: `go test ./internal/designsystem/`
Expected: `ok`.

---

### Task 5: Wire tokens, rules and items into `Load`

**Model:** haiku

**Files:**
- Modify: `internal/designsystem/load.go` (replace `Load`; append two helpers)
- Modify: `internal/designsystem/load_test.go` (append tests)

**Interfaces:**
- Consumes: `parseTokens` (Task 2), `parseRules` (Task 3), `loadItems` (Task 4), `ReadPointOfView` (slice 1).
- Produces: the final `Load(root string) (Pack, []Problem)`. Problems are reported in this order: point of view, tokens, rules, components, templates, flows.

- [ ] **Step 1: Write the failing tests**

Append to `internal/designsystem/load_test.go` (add nothing to the import block; the existing imports cover these tests):

```go
func TestLoadRequiredFilesMissing(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{"design-system/keep.txt": ""})
	_, problems := Load(root)
	checkProblems(t, problems, []wantProblem{
		{0, "file not found"}, // point of view
		{0, "file not found"}, // tokens
		{0, "file not found"}, // rules
	})
}

func TestLoadCountsEveryKind(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"design-system/point-of-view.md":     "# Brand\n",
		"design-system/tokens.json":          `{"color":{"ink":"#fff"}}`,
		"design-system/rules.md":             "## a\nS.\nReason: r\n",
		"design-system/components/button.md": "# Button\n",
		"design-system/templates/page.md":    "# Page\n",
		"design-system/flows/first-run.md":   "# First run\n",
	})
	pack, problems := Load(root)
	checkProblems(t, problems, nil)
	if pack.PointOfView == "" || len(pack.Tokens) != 1 || len(pack.Rules) != 1 ||
		len(pack.Components) != 1 || len(pack.Templates) != 1 || len(pack.Flows) != 1 {
		t.Fatalf("pack = %+v, want one of each kind", pack)
	}
}

func TestLoadReportsEveryBadFileAndKeepsTheRest(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"design-system/point-of-view.md":     "# Brand\n",
		"design-system/tokens.json":          "{\n  \"color\": {\"ink\": true}\n}\n",
		"design-system/rules.md":             "## Bad\nS.\nReason: r\n## good\nS.\nReason: r\n",
		"design-system/components/Button.md": "# Button\n",
	})
	pack, problems := Load(root)
	checkProblems(t, problems, []wantProblem{
		{2, `token "color.ink" must be a string or number`},
		{1, `rule id "Bad" must be`},
		{0, "file name must be lowercase kebab-case"},
	})
	if len(pack.Rules) != 1 || pack.Rules[0].ID != "good" {
		t.Errorf("rules = %+v, want only good", pack.Rules)
	}
	if pack.PointOfView == "" {
		t.Errorf("point of view was dropped because another file had problems")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/designsystem/ -run 'TestLoadRequiredFilesMissing|TestLoadCountsEveryKind|TestLoadReportsEveryBad' -v`
Expected: FAIL; the tokens, rules and items are not loaded yet, so the counts and problems do not match.

- [ ] **Step 3: Replace `Load` in `load.go` and append the helpers**

In `internal/designsystem/load.go`, replace the whole `Load` function with:

```go
// Load reads root/design-system and returns the pack it could read and every
// problem it found. A non-empty problems slice means the pack is not valid,
// but Load still returns whatever it could parse.
func Load(root string) (Pack, []Problem) {
	var pack Pack
	info, err := os.Stat(filepath.Join(root, DirName))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return pack, []Problem{{File: DirName, Message: "directory not found; run tenet init"}}
	case err != nil:
		return pack, []Problem{{File: DirName, Message: "could not read: " + err.Error()}}
	case !info.IsDir():
		return pack, []Problem{{File: DirName, Message: "not a directory"}}
	}

	var problems []Problem
	text, err := ReadPointOfView(root)
	switch {
	case errors.Is(err, ErrPointOfViewMissing):
		problems = append(problems, Problem{File: PointOfViewPath, Message: "file not found"})
	case errors.Is(err, ErrPointOfViewEmpty):
		problems = append(problems, Problem{File: PointOfViewPath, Line: 1, Message: "file is empty"})
	case err != nil:
		problems = append(problems, Problem{File: PointOfViewPath, Message: "could not read: " + err.Error()})
	default:
		pack.PointOfView = text
	}

	tokens, ps := loadTokens(root)
	pack.Tokens = tokens
	problems = append(problems, ps...)

	rules, ps := loadRules(root)
	pack.Rules = rules
	problems = append(problems, ps...)

	components, ps := loadItems(root, ComponentsDir)
	pack.Components = components
	problems = append(problems, ps...)

	templates, ps := loadItems(root, TemplatesDir)
	pack.Templates = templates
	problems = append(problems, ps...)

	flows, ps := loadItems(root, FlowsDir)
	pack.Flows = flows
	problems = append(problems, ps...)

	return pack, problems
}
```

Then append to the end of the same file:

```go
// loadTokens reads and parses the required tokens file.
func loadTokens(root string) ([]Token, []Problem) {
	data, problems := readRequired(root, TokensPath)
	if problems != nil {
		return nil, problems
	}
	return parseTokens(TokensPath, data)
}

// loadRules reads and parses the required rules file.
func loadRules(root string) ([]Rule, []Problem) {
	data, problems := readRequired(root, RulesPath)
	if problems != nil {
		return nil, problems
	}
	return parseRules(RulesPath, data)
}

// readRequired reads a required file. A missing or unreadable file becomes one problem.
func readRequired(root, rel string) ([]byte, []Problem) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, []Problem{{File: rel, Message: "file not found"}}
	case err != nil:
		return nil, []Problem{{File: rel, Message: "could not read: " + err.Error()}}
	}
	return data, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/designsystem/ -v`
Expected: PASS for all tests in the package, including the slice 1 `ReadPointOfView` tests.

- [ ] **Step 5: Run the full suite**

Run: `go test ./...`
Expected: `ok` for `cmd/tenet`, `internal/designsystem`, and `internal/server`. The server tests still pass because they only read the point of view.

---

### Task 6: `Scaffold` writes an empty, valid design system

**Model:** haiku

**Files:**
- Create: `internal/designsystem/scaffold.go`
- Create: `internal/designsystem/scaffold_test.go`

**Interfaces:**
- Consumes: `DirName`, `PointOfViewPath`, `TokensPath`, `RulesPath`, `ComponentsDir`, `TemplatesDir`, `FlowsDir`, `Load` (Task 5).
- Produces:
  - `var ErrAlreadyExists = errors.New("design-system already exists")`
  - `func Scaffold(root string) error`. If `root/design-system` exists in any form (file, directory, or symlink, including a dangling one), it returns an error wrapping `ErrAlreadyExists` and writes nothing. Otherwise it creates the folder and writes the files. If a write fails after the folder was created, it removes that folder and returns the error.

Scaffold content (D26):
- `design-system/point-of-view.md`: `# Point of view` heading, then a paragraph asking the author to replace it.
- `design-system/tokens.json`: `{}` followed by a newline.
- `design-system/rules.md`: `# Rules` heading, then one line explaining the entry format.
- `design-system/components/.gitkeep`, `templates/.gitkeep`, `flows/.gitkeep`: empty files.

- [ ] **Step 1: Write the failing test**

Create `internal/designsystem/scaffold_test.go`:

```go
package designsystem

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldWritesEveryFile(t *testing.T) {
	root := t.TempDir()
	if err := Scaffold(root); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	for _, rel := range []string{
		PointOfViewPath,
		TokensPath,
		RulesPath,
		ComponentsDir + "/.gitkeep",
		TemplatesDir + "/.gitkeep",
		FlowsDir + "/.gitkeep",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestScaffoldThenLoadIsValid(t *testing.T) {
	root := t.TempDir()
	if err := Scaffold(root); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	pack, problems := Load(root)
	checkProblems(t, problems, nil)
	if pack.PointOfView == "" {
		t.Errorf("point of view is empty in scaffold")
	}
	if len(pack.Tokens) != 0 || len(pack.Rules) != 0 || len(pack.Components) != 0 ||
		len(pack.Templates) != 0 || len(pack.Flows) != 0 {
		t.Errorf("scaffold is not empty: %+v", pack)
	}
}

func TestScaffoldRefusesExistingFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, DirName)
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Scaffold(root)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "x" {
		t.Errorf("existing file was changed to %q", data)
	}
}

func TestScaffoldRefusesExistingDirectory(t *testing.T) {
	root := t.TempDir()
	keep := filepath.Join(root, DirName, "keep.md")
	if err := os.MkdirAll(filepath.Dir(keep), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Scaffold(root); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
	if _, err := os.Stat(filepath.Join(root, PointOfViewPath)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("scaffold wrote into an existing folder: %v", err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Errorf("existing file was removed: %v", err)
	}
}

func TestScaffoldRefusesDanglingSymlink(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, DirName)
	if err := os.Symlink(filepath.Join(root, "nowhere"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := Scaffold(root); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
	if _, err := os.Lstat(link); err != nil {
		t.Errorf("symlink was removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "nowhere")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("scaffold wrote through the symlink: %v", err)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/designsystem/ -run TestScaffold -v`
Expected: FAIL with `undefined: Scaffold`.

- [ ] **Step 3: Write `scaffold.go`**

Create `internal/designsystem/scaffold.go`:

```go
package designsystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ErrAlreadyExists reports that design-system already exists, so Scaffold wrote nothing.
var ErrAlreadyExists = errors.New("design-system already exists")

const (
	pointOfViewScaffold = "# Point of view\n\nDescribe what this product stands for, who it is for, and what its users care about. Every rule should trace back to this.\n"
	tokensScaffold      = "{}\n"
	rulesScaffold       = "# Rules\n\nEach rule is a level-2 heading with its id, then a statement and a Reason: line.\n"
)

// scaffoldFiles maps each file Scaffold writes, relative to the project root, to its content.
var scaffoldFiles = map[string]string{
	PointOfViewPath:             pointOfViewScaffold,
	TokensPath:                  tokensScaffold,
	RulesPath:                   rulesScaffold,
	ComponentsDir + "/.gitkeep": "",
	TemplatesDir + "/.gitkeep":  "",
	FlowsDir + "/.gitkeep":      "",
}

// Scaffold writes an empty, valid design-system/ folder under root. It returns an
// error wrapping ErrAlreadyExists, and writes nothing, when the path exists in any
// form, including a dangling symlink.
func Scaffold(root string) error {
	dir := filepath.Join(root, DirName)
	if _, err := os.Lstat(dir); err == nil {
		return fmt.Errorf("%w under %s; nothing written", ErrAlreadyExists, root)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		return err
	}
	if err := writeScaffold(root); err != nil {
		_ = os.RemoveAll(dir)
		return err
	}
	return nil
}

// writeScaffold creates each scaffold file with exclusive create, so it never overwrites.
func writeScaffold(root string) error {
	for rel, content := range scaffoldFiles {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		if _, err := f.WriteString(content); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/designsystem/ -run TestScaffold -v`
Expected: PASS for all scaffold tests.

- [ ] **Step 5: Run the package tests**

Run: `go test ./internal/designsystem/`
Expected: `ok`.

---

### Task 7: Fixture pack for `testdata/project`

**Model:** sonnet

**Files:**
- Create: `testdata/project/design-system/tokens.json`
- Create: `testdata/project/design-system/rules.md`
- Create: `testdata/project/design-system/components/button.md`
- Create: `testdata/project/design-system/templates/settings-page.md`
- Create: `testdata/project/design-system/flows/first-run.md`
- Modify: `internal/designsystem/load_test.go` (append one test)

The existing `testdata/project/design-system/point-of-view.md` is unchanged.

**Interfaces:**
- Consumes: `Load` (Task 5).
- Produces: a fixture that loads with zero problems and these counts: 4 tokens, 2 rules, 1 component, 1 template, 1 flow. Used by the server tests (unchanged), the CLI tests in Tasks 8 and 9, and the README.

- [ ] **Step 1: Write the failing test**

Append to `internal/designsystem/load_test.go`:

```go
func TestLoadFixtureIsValid(t *testing.T) {
	pack, problems := Load("../../testdata/project")
	checkProblems(t, problems, nil)
	if len(pack.Tokens) != 4 {
		t.Errorf("tokens = %d, want 4", len(pack.Tokens))
	}
	if len(pack.Rules) != 2 {
		t.Errorf("rules = %d, want 2", len(pack.Rules))
	}
	if len(pack.Components) != 1 || len(pack.Templates) != 1 || len(pack.Flows) != 1 {
		t.Errorf("items = %d components, %d templates, %d flows; want 1 each",
			len(pack.Components), len(pack.Templates), len(pack.Flows))
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/designsystem/ -run TestLoadFixtureIsValid -v`
Expected: FAIL; the fixture has no tokens or rules yet, so Load reports `file not found` for tokens and rules.

- [ ] **Step 3: Create the fixture files with exactly this content**

`testdata/project/design-system/tokens.json`:

```json
{
  "color": {
    "ink": "#1d2433",
    "paper": "#fbfaf7",
    "accent": "#2f6fed"
  },
  "space": {
    "unit": 4
  }
}
```

`testdata/project/design-system/rules.md`:

````markdown
# Rules

## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.

## spacing.unit-multiples
Spacing between elements is a multiple of the space unit token.
Reason: Off-grid spacing is the quickest way for a screen to look unplanned.
````

`testdata/project/design-system/components/button.md`:

```markdown
# Button

The primary action on a surface. Use one per view.
```

`testdata/project/design-system/templates/settings-page.md`:

```markdown
# Settings page

A single column of labelled sections, each with its own save action.
```

`testdata/project/design-system/flows/first-run.md`:

```markdown
# First run

The steps a new user sees before the first screen: a welcome, one setup question, then the home page.
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/designsystem/ -run TestLoadFixtureIsValid -v`
Expected: PASS.

- [ ] **Step 5: Run the full suite, including the server tests that read this fixture**

Run: `go test ./...`
Expected: `ok` for all packages.

---

### Task 8: `tenet validate` command and subcommand dispatch

**Model:** sonnet

**Files:**
- Modify: `cmd/tenet/main.go` (full replacement below)
- Create: `cmd/tenet/validate.go`
- Modify: `cmd/tenet/main_test.go` (replace the import block; append tests)

**Interfaces:**
- Consumes: `designsystem.Load`, `designsystem.Pack`, `designsystem.Problem`, `parseRoot` (slice 1, unchanged).
- Produces:
  - `func run(args []string, cwd string, stdout, stderr io.Writer) int`. Replaces slice 1's `run(args, cwd, stderr)`. First argument `validate` runs `runValidate`. An unknown non-flag first argument exits 2 with `tenet: unknown command "<name>"`. Otherwise the serve path runs (`runServe`).
  - `func runServe(args []string, cwd string, stderr io.Writer) int`: slice 1's serve body, unchanged in this task.
  - `func runValidate(args []string, cwd string, stdout, stderr io.Writer) int`. Exit 0 when valid, 1 when there are problems, 2 for usage or `--dir` errors.
  - `func packSummary(p designsystem.Pack) string`: `N tokens, M rules, C components, T templates, F flows`.

- [ ] **Step 1: Write the failing tests**

Replace the import block of `cmd/tenet/main_test.go` with:

```go
import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)
```

Append to `cmd/tenet/main_test.go`:

```go
func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"nope"}, t.TempDir(), &stdout, &stderr); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), `unknown command "nope"`) {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestRunValidateFixture(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "--dir", "../../testdata/project"}, cwd, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	want := "design system valid: 4 tokens, 2 rules, 1 components, 1 templates, 1 flows"
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), want)
	}
}

func TestRunValidateReportsFileAndLine(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"design-system/point-of-view.md": "# Brand\n",
		"design-system/rules.md":         "# Rules\n",
		"design-system/tokens.json":      "{\n  \"color\": {\n    \"ink\": true\n  }\n}\n",
	})
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "--dir", root}, t.TempDir(), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	out := stdout.String()
	if !strings.Contains(out, `design-system/tokens.json:3: token "color.ink" must be a string or number`) {
		t.Errorf("stdout = %q, want a file:line problem", out)
	}
	if !strings.Contains(out, "1 problem(s) found") {
		t.Errorf("stdout = %q, want the problem count", out)
	}
}

func TestRunValidateMissingDesignSystem(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "--dir", t.TempDir()}, t.TempDir(), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "directory not found") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestRunValidateBadRoot(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "--dir", "does-not-exist"}, t.TempDir(), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

// writeFiles creates each file under root, making parent directories as needed.
func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./cmd/tenet/ -run 'TestRunUnknown|TestRunValidate' -v`
Expected: FAIL to compile, `too many arguments` for `run` and `undefined: runValidate`.

- [ ] **Step 3: Replace `cmd/tenet/main.go`**

Replace the whole file with:

```go
// Command tenet serves a project's design system to MCP clients over stdio, and
// checks (validate) or scaffolds (init) its design-system folder.
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
	"strings"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/server"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "tenet:", err)
		os.Exit(1)
	}
	os.Exit(run(os.Args[1:], cwd, os.Stdout, os.Stderr))
}

// run dispatches on the first argument and returns the process exit code. With no
// subcommand it serves MCP over stdio.
func run(args []string, cwd string, stdout, stderr io.Writer) int {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "validate":
			return runValidate(args[1:], cwd, stdout, stderr)
		default:
			fmt.Fprintf(stderr, "tenet: unknown command %q\n", args[0])
			return 2
		}
	}
	return runServe(args, cwd, stderr)
}

// runServe runs the stdio MCP server. It writes nothing to stdout.
func runServe(args []string, cwd string, stderr io.Writer) int {
	logger := log.New(stderr, "tenet: ", 0)
	root, err := parseRoot(args, cwd)
	if err != nil {
		logger.Println(err)
		return 2
	}
	logger.Printf("%s serving %s on stdio", server.Version, root)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.New(root).Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
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

- [ ] **Step 4: Create `cmd/tenet/validate.go`**

```go
package main

import (
	"fmt"
	"io"

	"github.com/tenet-ui/tenet/internal/designsystem"
)

// runValidate implements `tenet validate [--dir <path>]`. Exit codes: 0 valid,
// 1 problems found, 2 usage or --dir error.
func runValidate(args []string, cwd string, stdout, stderr io.Writer) int {
	root, err := parseRoot(args, cwd)
	if err != nil {
		fmt.Fprintln(stderr, "tenet validate:", err)
		return 2
	}
	pack, problems := designsystem.Load(root)
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(stdout, p)
		}
		fmt.Fprintf(stdout, "%d problem(s) found\n", len(problems))
		return 1
	}
	fmt.Fprintf(stdout, "design system valid: %s\n", packSummary(pack))
	return 0
}

// packSummary counts each kind of item in a pack, for output lines.
func packSummary(p designsystem.Pack) string {
	return fmt.Sprintf("%d tokens, %d rules, %d components, %d templates, %d flows",
		len(p.Tokens), len(p.Rules), len(p.Components), len(p.Templates), len(p.Flows))
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./cmd/tenet/ -v`
Expected: PASS for the new tests and the slice 1 `parseRoot` tests.

- [ ] **Step 6: Build and smoke-test the command**

Run: `go build -o /tmp/claude-1000/-home-hossein-Projects-github-com-tenet-ui-tenet/db1039aa-c355-474a-817f-567d4adb80e4/scratchpad/tenet-check ./cmd/tenet && /tmp/claude-1000/-home-hossein-Projects-github-com-tenet-ui-tenet/db1039aa-c355-474a-817f-567d4adb80e4/scratchpad/tenet-check validate --dir testdata/project; echo "exit=$?"`

The binary goes to the session scratchpad, not the repo, so it is not left in the working tree.

Expected: the line `design system valid: 4 tokens, 2 rules, 1 components, 1 templates, 1 flows` and `exit=0`.

---

### Task 9: `tenet init` command

**Model:** sonnet

**Files:**
- Create: `cmd/tenet/init.go`
- Create: `cmd/tenet/init_test.go`
- Modify: `cmd/tenet/main.go` (one case added to the switch in `run`)

**Interfaces:**
- Consumes: `designsystem.Scaffold`, `designsystem.ErrAlreadyExists`, `parseRoot`, `runValidate` (Task 8).
- Produces: `func runInit(args []string, cwd string, stdout, stderr io.Writer) int`. Exit 0 when written, 1 when the folder already exists or a write fails, 2 for usage or `--dir` errors.

- [ ] **Step 1: Write the failing test**

Create `cmd/tenet/init_test.go`:

```go
package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunInitScaffoldsAValidFolder(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", "--dir", root}, t.TempDir(), &stdout, &stderr); code != 0 {
		t.Fatalf("init exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "created design-system/") {
		t.Errorf("stdout = %q", stdout.String())
	}

	stdout.Reset()
	if code := run([]string{"validate", "--dir", root}, t.TempDir(), &stdout, &stderr); code != 0 {
		t.Fatalf("validate exit = %d, stdout = %q", code, stdout.String())
	}
	want := "design system valid: 0 tokens, 0 rules, 0 components, 0 templates, 0 flows"
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), want)
	}
}

func TestRunInitRefusesExistingFolder(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", "--dir", root}, t.TempDir(), &stdout, &stderr); code != 0 {
		t.Fatalf("first init exit = %d", code)
	}
	stderr.Reset()
	if code := run([]string{"init", "--dir", root}, t.TempDir(), &stdout, &stderr); code != 1 {
		t.Fatalf("second init exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "already exists") {
		t.Errorf("stderr = %q, want it to say the folder already exists", stderr.String())
	}
}

func TestRunInitBadRoot(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", "--dir", "does-not-exist"}, t.TempDir(), &stdout, &stderr); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/tenet/ -run TestRunInit -v`
Expected: FAIL: `init` is an unknown command, so exit is 2 and the first assertion fails.

- [ ] **Step 3: Create `cmd/tenet/init.go`**

```go
package main

import (
	"fmt"
	"io"

	"github.com/tenet-ui/tenet/internal/designsystem"
)

// runInit implements `tenet init [--dir <path>]`. Exit codes: 0 written, 1 the
// folder already exists or the write failed, 2 usage or --dir error.
func runInit(args []string, cwd string, stdout, stderr io.Writer) int {
	root, err := parseRoot(args, cwd)
	if err != nil {
		fmt.Fprintln(stderr, "tenet init:", err)
		return 2
	}
	if err := designsystem.Scaffold(root); err != nil {
		fmt.Fprintln(stderr, "tenet init:", err)
		return 1
	}
	fmt.Fprintf(stdout, "created design-system/ under %s; next, write design-system/point-of-view.md\n", root)
	return 0
}
```

- [ ] **Step 4: Add the `init` case to `run` in `cmd/tenet/main.go`**

In `run`, change the switch so it reads:

```go
		switch args[0] {
		case "validate":
			return runValidate(args[1:], cwd, stdout, stderr)
		case "init":
			return runInit(args[1:], cwd, stdout, stderr)
		default:
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./cmd/tenet/ -v`
Expected: PASS for `TestRunInit*` and all earlier tests.

---

### Task 10: Serve path logs the pack summary and problems at startup

**Model:** haiku

**Files:**
- Modify: `cmd/tenet/main.go` (`runServe` only)

**Interfaces:**
- Consumes: `designsystem.Load`, `packSummary` (Task 8).
- Produces: the same serve behaviour, plus stderr lines at startup. Nothing is written to stdout.

Startup output (D18, D19, spec "Module interfaces"):
- `tenet <Version> serving <root> on stdio; design system: <packSummary>`
- If there are problems: `design system has K problem(s); run tenet validate for details`, then one `tenet: <problem>` line per problem.

- [ ] **Step 1: Make the change**

In `cmd/tenet/main.go`, replace this block inside `runServe`:

```go
	logger.Printf("%s serving %s on stdio", server.Version, root)
```

with:

```go
	pack, problems := designsystem.Load(root)
	logger.Printf("%s serving %s on stdio; design system: %s", server.Version, root, packSummary(pack))
	if len(problems) > 0 {
		logger.Printf("design system has %d problem(s); run tenet validate for details", len(problems))
		for _, p := range problems {
			logger.Println(p)
		}
	}
```

Then add the import `"github.com/tenet-ui/tenet/internal/designsystem"` to the import block of `cmd/tenet/main.go`, in the same group as `internal/server`.

- [ ] **Step 2: Verify it builds and the test suite still passes**

Run: `go test ./...`
Expected: `ok` for all packages.

- [ ] **Step 3: Verify the stdio check still produces only protocol on stdout**

Run the README's stdio check (Task 11 gives the exact commands) against `testdata/project` and confirm stdout holds only JSON-RPC lines and the startup line appears on stderr.

Expected: stderr shows `tenet 0.1.0 serving ... on stdio; design system: 4 tokens, 2 rules, 1 components, 1 templates, 1 flows`; stdout shows the initialize and tools/call responses only.

---

### Task 11: README covers the format and the commands

**Model:** haiku

**Files:**
- Modify: `README.md` (full replacement below)

**Interfaces:**
- Consumes: the commands and format from Tasks 2–10.
- Produces: user documentation. Nothing else depends on it.

- [ ] **Step 1: Replace `README.md` with exactly this content**

````markdown
# tenet

An MCP server that lets Claude users encode their own design system (point of view, tokens, components, templates, flows, rules) and delivers it to the AI at the right moment, so AI-built interfaces carry intent instead of generic "zombie UI". The Stripe design system, reverse-engineered from public sources, is the reference example.

The tool has three commands: serving MCP over stdio (the default), `validate`, and `init`. It has one MCP tool, `get_point_of_view`. The design-system tools arrive in a later slice.

## Requirements

Go 1.27.0 (the toolchain version in `go.mod`).

## Run

```bash
go run ./cmd/tenet --dir testdata/project
go build -o tenet ./cmd/tenet
```

Without `--dir`, the working directory is the project root.

## Start a design system

```bash
go run ./cmd/tenet init --dir /path/to/your/project
```

This writes an empty, valid `design-system/` folder. It refuses to run if `design-system` already exists, and it writes nothing in that case.

## Check a design system

```bash
go run ./cmd/tenet validate --dir testdata/project
```

Each problem prints as `path:line: message`, so most terminals and editors make it a link. The command exits 0 when the system is valid, 1 when it has problems, and 2 for a usage or `--dir` error.

## The design-system folder

```
design-system/
├── point-of-view.md     required: the brand point of view, free Markdown, not empty
├── tokens.json          required: design values, see below
├── rules.md             required: checkable rules, see below
├── components/          optional: one Markdown file per component
├── templates/           optional: one Markdown file per template
└── flows/               optional: one Markdown file per flow
```

**`tokens.json`** is an object of categories. Each category is an object of token names. A value is a string or a number:

```json
{
  "color": {
    "ink": "#1d2433",
    "accent": "#2f6fed"
  },
  "space": {
    "unit": 4
  }
}
```

Category and token names are lowercase letters, digits and dashes, starting with a letter (`ink`, `touch-target`). A token's full name is `category.name`, for example `color.ink`.

**`rules.md`** holds one rule per `##` heading. The heading is the rule's id, lowercase letters and digits joined by dots or dashes. Each rule has a statement and exactly one `Reason:` line:

```markdown
## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
```

A rules file with no rules is valid.

**`components/`, `templates/`, `flows/`** each hold Markdown files. The file name, without `.md`, is the item's id and must be lowercase kebab-case (`primary-button.md`). The first line must be a level-1 heading with the item's title:

```markdown
# Button

The primary action on a surface. Use one per view.
```

Files in those folders that do not end in `.md` are ignored.

See [testdata/project/design-system](testdata/project/design-system) for a complete example that passes `validate`.

## Check the stdio protocol

```bash
( printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"check","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_point_of_view","arguments":{}}}'; \
  sleep 3 ) | go run ./cmd/tenet --dir testdata/project
```

Stdin must stay open until the responses arrive, which is why the command sleeps. Stdout carries only JSON-RPC responses. The startup line, which includes the design-system summary and any problems, goes to stderr.

## Connect Claude Code

```bash
claude mcp add --transport stdio tenet -- /absolute/path/to/tenet --dir /absolute/path/to/your/project
```

Then run `claude mcp list` and confirm `tenet` is connected. Run `tenet validate --dir /absolute/path/to/your/project` first, so the point of view and the other files are known to be valid.

## Develop

```bash
go test ./...
test -z "$(gofmt -l .)"
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
```

## Docs

- [GLOSSARY.md](GLOSSARY.md)
- [docs/roadmap.md](docs/roadmap.md)
- [docs/adr/](docs/adr/)
````

- [ ] **Step 2: Verify the README examples still hold**

Run: `go run ./cmd/tenet validate --dir testdata/project; echo "exit=$?"`
Expected: `design system valid: 4 tokens, 2 rules, 1 components, 1 templates, 1 flows` and `exit=0`.

---

## Final verification (for slice-build B5, not a plan task)

Run these after Task 11, and read the output before claiming success:

1. `go test ./...`
2. `test -z "$(gofmt -l .)"`
3. `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`
4. `go run ./cmd/tenet validate --dir testdata/project` exits 0
5. A bad fixture exits 1 with `file:line`: copy `testdata/project` to a temp directory, change one token value to `true` on a known line, and run validate
6. `tenet init` into an empty temp directory, then `tenet validate` exits 0; a second `tenet init` exits 1
7. The README stdio check returns the point of view on stdout, with the startup summary on stderr
8. Acceptance criteria 1–3 of the slice, each observed in the running binary
