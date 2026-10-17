# Slice 4: Critique tool Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `critique` MCP tool that checks one UI source file against the rules of a design system and returns one finding per breach, each naming the rule, the line, the matched text and a specific fix.

**Architecture:** Rules gain two optional body lines, `Forbid:` (a Go RE2 pattern) and `Fix:` (text), parsed in `designsystem/rules.go`. A pure `Critique` function in `designsystem` matches the patterns line by line. The server reads the source file, loads the chosen system, and renders the findings through one function. Detection is deterministic because the server has no model access; judgement stays with the client (ADR-0009, written in Task 5).

**Tech Stack:** Go 1.27 (`go.mod`), `github.com/modelcontextprotocol/go-sdk` v1.8.0, standard library only in `internal/designsystem` (`regexp`, `strings`, `bytes`).

**Spec:** `docs/superpowers/specs/2026-10-18-slice-4-critique-tool-design.md` (approved, commit 54cd1c7). Decision numbers below (C1–C22) refer to that spec.

## Global Constraints

- Module `github.com/tenet-ui/tenet`; Go 1.27; go-sdk `v1.8.0`. No new dependencies.
- `internal/designsystem` imports only the standard library and has no MCP types.
- While serving, stdout carries only MCP protocol; logs go to stderr.
- Tool failures are `IsError` tool results, not Go errors (slice 1 convention).
- Patterns are Go RE2 (`regexp` package). Source files are at most 1 MiB (`1 << 20` bytes).
- Test command for everything: `go test ./...`; format: `test -z "$(gofmt -l .)"`; lint: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`.
- Do not add commit steps. Implementers and fixers do not commit; the orchestrator makes one commit at the end.

## Review Focus

1. A `Forbid:` pattern that matches the empty string (for example `x*`) → no empty findings. Test: Task 2 `TestCritiqueIgnoresEmptyMatches`; Task 3 `TestCritiqueNoFindings`.
2. A source file with CRLF line endings → matched text never includes `\r`. Test: Task 2 `TestCritiqueCRLF`.
3. A source file with a UTF-8 BOM → the BOM is stripped, so `^` anchors match line 1. Test: Task 2 `TestCritiqueStripsBOM`.
4. A binary file (NUL byte) passed as `file` → an error, not findings computed from binary data. Test: Task 3 `TestCritiqueBinaryFileIsError`.
5. A `Forbid:` pattern that does not compile → the rule is invalid, `validate` reports it, and the rule is not silently checked. Tests: Task 1 `TestParseRulesCheck` (invalid pattern case); Task 3 `TestCritiqueRulesProblemNote`.

---

### Task 1: Parse `Forbid:` and `Fix:` lines in rules.md

**Files:**
- Modify: `internal/designsystem/rules.go`
- Test: `internal/designsystem/rules_test.go`

**Interfaces:**
- Consumes: the existing `parseRules(file string, data []byte) ([]Rule, []Problem)` and the `add` closure, `wantProblem`, `checkProblems`, `RulesPath` already in the package.
- Produces: `Rule.Forbid *regexp.Regexp` (nil when the rule has no pattern) and `Rule.Fix string` (empty when none). Consumed by Task 2 (`Critique`, `Checkable`).

- [ ] **Step 1: Write the failing test**

Append to `internal/designsystem/rules_test.go`:

```go
func TestParseRulesCheck(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantValid    bool
		wantForbid   string // "" when Forbid is nil
		wantFix      string
		wantProblems []wantProblem
	}{
		{
			name:      "no Forbid or Fix is valid",
			data:      "## a\nS.\nReason: r.\n",
			wantValid: true,
		},
		{
			name:       "Forbid with Fix",
			data:       "## a\nS.\nReason: r.\nForbid: #[0-9]+\nFix: Use a token.\n",
			wantValid:  true,
			wantForbid: `#[0-9]+`,
			wantFix:    "Use a token.",
		},
		{
			name:         "Forbid without Fix",
			data:         "## a\nS.\nReason: r.\nForbid: x\n",
			wantProblems: []wantProblem{{4, `rule "a" has a Forbid: line without a Fix: line`}},
		},
		{
			name:         "Fix without Forbid",
			data:         "## a\nS.\nReason: r.\nFix: Use a token.\n",
			wantProblems: []wantProblem{{4, `rule "a" has a Fix: line without a Forbid: line`}},
		},
		{
			name:         "repeated Forbid",
			data:         "## a\nS.\nReason: r.\nForbid: x\nForbid: y\nFix: f\n",
			wantProblems: []wantProblem{{5, `rule "a" has more than one Forbid: line`}},
		},
		{
			name:         "repeated Fix",
			data:         "## a\nS.\nReason: r.\nForbid: x\nFix: f\nFix: g\n",
			wantProblems: []wantProblem{{6, `rule "a" has more than one Fix: line`}},
		},
		{
			name:         "empty Forbid",
			data:         "## a\nS.\nReason: r.\nForbid:\nFix: f\n",
			wantProblems: []wantProblem{{4, `rule "a" has an empty Forbid: line`}},
		},
		{
			name:         "empty Fix",
			data:         "## a\nS.\nReason: r.\nForbid: x\nFix:\n",
			wantProblems: []wantProblem{{5, `rule "a" has an empty Fix: line`}},
		},
		{
			name:         "invalid pattern",
			data:         "## a\nS.\nReason: r.\nForbid: (\nFix: f\n",
			wantProblems: []wantProblem{{4, `rule "a" has an invalid Forbid: pattern`}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, problems := parseRules(RulesPath, []byte(tt.data))
			checkProblems(t, problems, tt.wantProblems)
			if len(rules) != 1 {
				t.Fatalf("got %d rules, want 1", len(rules))
			}
			r := rules[0]
			if r.Valid != tt.wantValid {
				t.Errorf("Valid = %t, want %t", r.Valid, tt.wantValid)
			}
			gotForbid := ""
			if r.Forbid != nil {
				gotForbid = r.Forbid.String()
			}
			if gotForbid != tt.wantForbid {
				t.Errorf("Forbid = %q, want %q", gotForbid, tt.wantForbid)
			}
			if r.Fix != tt.wantFix {
				t.Errorf("Fix = %q, want %q", r.Fix, tt.wantFix)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/designsystem -run TestParseRulesCheck -v`
Expected: FAIL to compile with `r.Forbid undefined` (the `Rule` struct has no `Forbid` or `Fix` field yet).

- [ ] **Step 3: Write minimal implementation**

In `internal/designsystem/rules.go`:

1. Add `"regexp"` to the import block (it already imports `fmt`, `regexp`, `strings`; `regexp` is present, so no change if already there — check first).

2. In `internal/designsystem/rules.go`, the `Rule` type is defined in `designsystem.go`'s neighbour `load.go` (the struct lives where `Rule` is declared — `internal/designsystem/load.go`). Add two fields to `type Rule struct` in `internal/designsystem/load.go`, after `Valid bool`:

```go
	Forbid *regexp.Regexp // nil when the rule has no Forbid: line
	Fix    string         // text of the Fix: line; empty when none
```
Add `"regexp"` to `load.go`'s import block.

3. In `parseRules`, declare these variables next to `refs []Ref` inside the `var (` block of each entry:

```go
			pattern    string
			forbidLine int
			forbids    int
			fix        string
			fixLine    int
			fixes      int
```

4. In the body loop, add these two branches immediately before the `if t != ""` branch (after the `Applies to:` branch):

```go
			if strings.HasPrefix(t, "Forbid:") {
				forbids++
				if forbids > 1 {
					add(l.no, "rule %q has more than one Forbid: line", e.id)
					continue
				}
				pattern = strings.TrimSpace(strings.TrimPrefix(t, "Forbid:"))
				forbidLine = l.no
				continue
			}
			if strings.HasPrefix(t, "Fix:") {
				fixes++
				if fixes > 1 {
					add(l.no, "rule %q has more than one Fix: line", e.id)
					continue
				}
				fix = strings.TrimSpace(strings.TrimPrefix(t, "Fix:"))
				fixLine = l.no
				continue
			}
```

5. After the `Reason:` checks (after the `if reasons == 0 ... else if reason == ""` block) and before `rules = append(...)`, add:

```go
		var forbid *regexp.Regexp
		switch {
		case forbids == 0 && fixes == 0:
		case forbids == 0:
			add(fixLine, "rule %q has a Fix: line without a Forbid: line", e.id)
		case fixes == 0:
			add(forbidLine, "rule %q has a Forbid: line without a Fix: line", e.id)
		case pattern == "":
			add(forbidLine, "rule %q has an empty Forbid: line", e.id)
		case fix == "":
			add(fixLine, "rule %q has an empty Fix: line", e.id)
		default:
			re, err := regexp.Compile(pattern)
			if err != nil {
				add(forbidLine, "rule %q has an invalid Forbid: pattern: %v", e.id, err)
			} else {
				forbid = re
			}
		}
```

6. In the `rules = append(rules, Rule{...})` literal, add two fields after `Valid:`... (order in struct literal does not matter; add after `AppliesTo: refs,`):

```go
			Forbid:    forbid,
			Fix:       fix,
```

Note: `Valid` is computed as `len(problems) == before`, so the new problems make the rule invalid automatically.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/designsystem -v`
Expected: PASS for `TestParseRulesCheck` and all existing tests (the existing `TestParseRules` expectations are unchanged because `formatRules` does not read the new fields).

Run: `gofmt -l internal/designsystem`
Expected: no output.

**Model:** sonnet (ordering of problem cases and the struct edits across two files need care).

---

### Task 2: Pure `Critique` and `Checkable` in designsystem

**Files:**
- Create: `internal/designsystem/critique.go`
- Test: `internal/designsystem/critique_test.go`

**Interfaces:**
- Consumes: `Rule` with `Valid`, `ID`, `Forbid`, `Fix` (Task 1); `stripBOM([]byte) []byte` from `problem.go`.
- Produces:
  - `type Finding struct { Line int; RuleID string; Match string; Fix string }`
  - `func Checkable(rules []Rule) int`
  - `func Critique(source string, rules []Rule) []Finding`
  Consumed by Task 3.

- [ ] **Step 1: Write the failing test**

Create `internal/designsystem/critique_test.go`:

```go
package designsystem

import (
	"reflect"
	"regexp"
	"testing"
)

// rule builds a valid checkable rule for the critique tests.
func rule(id, pattern, fix string) Rule {
	return Rule{ID: id, Valid: true, Forbid: regexp.MustCompile(pattern), Fix: fix}
}

func TestCritique(t *testing.T) {
	tests := []struct {
		name   string
		source string
		rules  []Rule
		want   []Finding
	}{
		{
			name:   "two matches on one line are two findings in match order",
			source: "a a",
			rules:  []Rule{rule("r", "a", "F.")},
			want: []Finding{
				{Line: 1, RuleID: "r", Match: "a", Fix: "F."},
				{Line: 1, RuleID: "r", Match: "a", Fix: "F."},
			},
		},
		{
			name:   "findings sort by line, then rule order",
			source: "a b\nb",
			rules:  []Rule{rule("a", "b", "Fa."), rule("b", "a", "Fb.")},
			want: []Finding{
				{Line: 1, RuleID: "a", Match: "b", Fix: "Fa."},
				{Line: 1, RuleID: "b", Match: "a", Fix: "Fb."},
				{Line: 2, RuleID: "a", Match: "b", Fix: "Fa."},
			},
		},
		{
			name:   "invalid rule is skipped",
			source: "a",
			rules:  []Rule{{ID: "x", Valid: false, Forbid: regexp.MustCompile("a"), Fix: "F."}},
			want:   nil,
		},
		{
			name:   "rule without Forbid is skipped",
			source: "a",
			rules:  []Rule{{ID: "x", Valid: true}},
			want:   nil,
		},
		{
			name:   "no rules, no findings",
			source: "a",
			rules:  nil,
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Critique(tt.source, tt.rules)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Critique() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCritiqueIgnoresEmptyMatches(t *testing.T) {
	got := Critique("abc", []Rule{rule("r", "x*", "F.")})
	if len(got) != 0 {
		t.Errorf("Critique() = %+v, want no findings", got)
	}
}

func TestCritiqueCRLF(t *testing.T) {
	got := Critique("a\r\nb\r\n", []Rule{rule("r", "a$", "F.")})
	want := []Finding{{Line: 1, RuleID: "r", Match: "a", Fix: "F."}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Critique() = %+v, want %+v", got, want)
	}
}

func TestCritiqueStripsBOM(t *testing.T) {
	got := Critique("﻿a", []Rule{rule("r", "^a", "F.")})
	want := []Finding{{Line: 1, RuleID: "r", Match: "a", Fix: "F."}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Critique() = %+v, want %+v", got, want)
	}
}

func TestCheckable(t *testing.T) {
	rules := []Rule{
		rule("a", "x", "F."),
		{ID: "b", Valid: true},
		{ID: "c", Valid: false, Forbid: regexp.MustCompile("x"), Fix: "F."},
	}
	if got := Checkable(rules); got != 1 {
		t.Errorf("Checkable() = %d, want 1", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/designsystem -run 'TestCritique|TestCheckable' -v`
Expected: FAIL to compile with `undefined: Critique` (and `Checkable`, `Finding`).

- [ ] **Step 3: Write minimal implementation**

Create `internal/designsystem/critique.go`:

```go
package designsystem

import "strings"

// Finding is one breach of one rule in one source file.
type Finding struct {
	Line   int    // 1-based line in the source
	RuleID string // the rule that was broken
	Match  string // the text the rule's pattern matched
	Fix    string // the rule's Fix: text
}

// Checkable returns the number of valid rules that have a Forbid pattern.
func Checkable(rules []Rule) int {
	n := 0
	for _, r := range rules {
		if r.Valid && r.Forbid != nil {
			n++
		}
	}
	return n
}

// Critique matches the Forbid pattern of each checkable rule against each line of source.
// Lines are split on "\n" and a trailing "\r" is removed from each. A leading UTF-8 BOM is
// removed first. Zero-length matches are ignored. Findings are in line order, then rule
// order, then match order.
func Critique(source string, rules []Rule) []Finding {
	text := string(stripBOM([]byte(source)))
	var findings []Finding
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSuffix(line, "\r")
		for _, r := range rules {
			if !r.Valid || r.Forbid == nil {
				continue
			}
			for _, m := range r.Forbid.FindAllString(line, -1) {
				if m == "" {
					continue
				}
				findings = append(findings, Finding{Line: i + 1, RuleID: r.ID, Match: m, Fix: r.Fix})
			}
		}
	}
	return findings
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/designsystem -v`
Expected: PASS for all tests in the package.

Run: `gofmt -l internal/designsystem`
Expected: no output.

**Model:** haiku (two files, exact paths, the full code is given, and the test cases are listed with their expected values).

---

### Task 3: The `critique` tool, its fixture, and the rules that use it

**Files:**
- Create: `internal/server/critique.go`
- Create: `internal/server/critique_test.go`
- Modify: `internal/server/server.go` (register the tool)
- Modify: `internal/server/server_test.go` (tool list is six tools)
- Modify: `testdata/project/design-system/rules.md` (add `Forbid:` and `Fix:` lines)
- Create: `testdata/critique/page.html` (fixture with two seeded violations)

**Interfaces:**
- Consumes: `designsystem.Load`, `designsystem.Critique`, `designsystem.Checkable`, `designsystem.Finding`, `designsystem.Problem`, `designsystem.RulesPath` (Tasks 1–2 and existing code); `loadPack`, `problemFor`, `textResult`, `toolError` from `internal/server/delivery.go`; `connect`, `callTool`, `textOf`, `writeDesignSystem`, `fixtureRoot` from the existing tests.
- Produces: tool `critique` with input `CritiqueInput{File, System string}`; handler `critiqueTool(root string, in CritiqueInput) *mcp.CallToolResult`.

- [ ] **Step 1: Write the failing test**

Create `internal/server/critique_test.go`:

```go
package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// critiqueRoot is the testdata folder: it holds critique/page.html and project/design-system.
const critiqueRoot = "../../testdata"

// wantFixtureText is the exact result for the fixture against the bundled example.
const wantFixtureText = `2 finding(s) in critique/page.html against project.

critique/page.html:4: color.tokens-only: found "#ff0000". Replace the raw colour with a colour token from tokens.json.
critique/page.html:5: spacing.unit-multiples: found "margin: 13px". Replace the raw pixel value with a space token from tokens.json, so spacing stays a multiple of the space unit.
`

func TestCritiqueFixtureFindings(t *testing.T) {
	session := connect(t, critiqueRoot)
	res := callTool(t, session, "critique", map[string]any{"file": "critique/page.html", "system": "project"})
	if res.IsError || textOf(t, res) != wantFixtureText {
		t.Errorf("got IsError=%t\n%s\nwant\n%s", res.IsError, textOf(t, res), wantFixtureText)
	}
}

func TestCritiqueNoFindings(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: x*\nFix: F.\n",
		"ui/page.html":                   "clean\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	want := "No findings in ui/page.html against . Checked 1 rule(s).\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueNoCheckableRules(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\n",
		"ui/page.html":                   "x\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	want := "No rule in . has a Forbid: line, so nothing was checked. Add one under a rule in design-system/rules.md.\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueRulesProblemNote(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: (\nFix: F.\n",
		"ui/page.html":                   "x\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	want := "No rule in . has a Forbid: line, so nothing was checked. Add one under a rule in design-system/rules.md.\n\n" +
		"Note: design-system/rules.md has problems, so some rules may not be checked. Run tenet validate.\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueCRLFAndBOM(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: ^a$\nFix: F.\n",
		"ui/page.html":                   "﻿a\r\nb\r\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	want := "1 finding(s) in ui/page.html against .\n\nui/page.html:1: r: found \"a\". F.\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueMissingFileIsError(t *testing.T) {
	session := connect(t, critiqueRoot)
	res := callTool(t, session, "critique", map[string]any{"file": "nope.html", "system": "project"})
	want := `file "nope.html" not found.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueDirectoryIsError(t *testing.T) {
	session := connect(t, critiqueRoot)
	res := callTool(t, session, "critique", map[string]any{"file": "critique", "system": "project"})
	want := `file "critique" is a directory; pass a file.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueBinaryFileIsError(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: a\nFix: F.\n",
	})
	if err := os.WriteFile(filepath.Join(root, "bin.dat"), []byte("a\x00a"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "bin.dat", "system": "."})
	want := `file "bin.dat" is not a text file.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueTooLargeIsError(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: a\nFix: F.\n",
	})
	big := strings.Repeat("a", 1<<20+1)
	if err := os.WriteFile(filepath.Join(root, "big.txt"), []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "big.txt", "system": "."})
	want := `file "big.txt" is larger than 1 MiB; pass a smaller source file.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueNoDesignSystemIsError(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "page.html"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "page.html", "system": "."})
	if !res.IsError || !strings.Contains(textOf(t, res), "run tenet init") {
		t.Errorf("got IsError=%t %q, want an error naming tenet init", res.IsError, textOf(t, res))
	}
}
```

Also in `internal/server/server_test.go`, rename `TestListsFiveTools` to `TestListsSixTools` and change its `want` to `[]string{"critique", "get_component", "get_flow", "get_point_of_view", "get_template", "get_tokens"}`.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/server -run 'Critique|TestListsSixTools' -v`
Expected: FAIL. `TestListsSixTools` fails (five tools listed), and the critique tests fail with `unknown tool "critique"` or a missing-fixture error.

- [ ] **Step 3: Write minimal implementation**

**3a. Fixture.** Create `testdata/critique/page.html` with exactly this content (line 4 and line 5 are the two seeded violations):

```html
<!doctype html>
<html>
<head><style>
.card { color: var(--color-ink); margin: var(--space-4); }
.hero { color: #ff0000; }
.note { margin: 13px; }
</style></head>
<body><p class="card">Hello</p></body>
</html>
```

**3b. Rules.** In `testdata/project/design-system/rules.md`, add two lines to the `color.tokens-only` entry and two lines to the `spacing.unit-multiples` entry. The entries become:

```markdown
## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
Applies to: tokens:color, component:button
Forbid: #[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?)\(
Fix: Replace the raw colour with a colour token from tokens.json.

## spacing.unit-multiples
Spacing between elements is a multiple of the space unit token.
Reason: Off-grid spacing is the quickest way for a screen to look unplanned.
Applies to: tokens:space, template:settings-page
Forbid: \b(margin|padding|gap)\b[^;"]*\b\d+px
Fix: Replace the raw pixel value with a space token from tokens.json, so spacing stays a multiple of the space unit.
```

**3c. Server handler.** Create `internal/server/critique.go`:

```go
package server

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/designsystem"
)

const (
	critiqueMaxBytes   = 1 << 20
	critiqueBinaryScan = 8000
	critiqueRulesNote  = "Note: design-system/rules.md has problems, so some rules may not be checked. Run tenet validate."
)

// CritiqueInput is the input of critique.
type CritiqueInput struct {
	File   string `json:"file" jsonschema:"path to the UI source file, relative to the project root or absolute"`
	System string `json:"system" jsonschema:"folder holding the design-system/ to check against, relative to the project root; . for the project root"`
}

// critiqueTool checks one source file against the checkable rules of one design system.
func critiqueTool(root string, in CritiqueInput) *mcp.CallToolResult {
	path := in.File
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	info, err := os.Stat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return toolError(fmt.Sprintf("file %q not found.", in.File))
	case err != nil:
		return toolError(fmt.Sprintf("could not read file %q: %v", in.File, err))
	case info.IsDir():
		return toolError(fmt.Sprintf("file %q is a directory; pass a file.", in.File))
	case info.Size() > critiqueMaxBytes:
		return toolError(fmt.Sprintf("file %q is larger than 1 MiB; pass a smaller source file.", in.File))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return toolError(fmt.Sprintf("could not read file %q: %v", in.File, err))
	}
	if bytes.IndexByte(data[:min(len(data), critiqueBinaryScan)], 0) >= 0 {
		return toolError(fmt.Sprintf("file %q is not a text file.", in.File))
	}

	systemRoot := in.System
	if !filepath.IsAbs(systemRoot) {
		systemRoot = filepath.Join(root, systemRoot)
	}
	pack, problems, errRes := loadPack(systemRoot)
	if errRes != nil {
		return errRes
	}
	findings := designsystem.Critique(string(data), pack.Rules)
	return textResult(renderCritique(in.File, in.System, findings, designsystem.Checkable(pack.Rules), rulesProblem(problems)))
}

// rulesProblem reports whether rules.md has a problem, so the result should carry the rules note.
func rulesProblem(problems []designsystem.Problem) bool {
	_, ok := problemFor(problems, designsystem.RulesPath)
	return ok
}

// renderCritique builds the text of a critique result. With findings it is a header, a blank
// line and one line per finding. Without findings it says how many rules were checked, or that
// none could be. The rules note follows a blank line when rules.md has problems.
func renderCritique(file, system string, findings []designsystem.Finding, checkable int, note bool) string {
	var lines []string
	switch {
	case checkable == 0:
		lines = []string{fmt.Sprintf("No rule in %s has a Forbid: line, so nothing was checked. Add one under a rule in design-system/rules.md.", system)}
	case len(findings) == 0:
		lines = []string{fmt.Sprintf("No findings in %s against %s. Checked %d rule(s).", file, system, checkable)}
	default:
		lines = []string{fmt.Sprintf("%d finding(s) in %s against %s.", len(findings), file, system), ""}
		for _, f := range findings {
			lines = append(lines, fmt.Sprintf("%s:%d: %s: found \"%s\". %s", file, f.Line, f.RuleID, f.Match, f.Fix))
		}
	}
	text := strings.Join(lines, "\n") + "\n"
	if note {
		text += "\n" + critiqueRulesNote + "\n"
	}
	return text
}
```

**3d. Register the tool.** In `internal/server/server.go`, inside `New`, after the `get_point_of_view` registration and before the loop over delivery tools, add:

```go
	mcp.AddTool(s, &mcp.Tool{
		Name:        "critique",
		Description: "Checks one UI source file against the design system's rules that have a Forbid: pattern. Returns one finding per match, naming the rule, the line, the matched text and the fix. Pass file (a path to the source file) and system (the folder holding design-system/; . for the project root).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in CritiqueInput) (*mcp.CallToolResult, any, error) {
		return critiqueTool(root, in), nil, nil
	})
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/server -v`
Expected: PASS for every critique test, `TestListsSixTools`, and all existing server tests (including `TestResultsUnder4KB` and the delivery tests, whose output does not change because `Forbid:`/`Fix:` are not part of the statement).

Run: `go test ./...`
Expected: all packages PASS.

Run: `gofmt -l .` and `go vet ./...`
Expected: no output from gofmt; vet clean.

Run: `go run ./cmd/tenet validate --dir testdata/project`
Expected: exit 0 (the new lines are valid).

**Model:** sonnet (four files plus fixtures; handler error ordering and exact rendered text need care).

---

### Task 4: Document the format and the tool in the README

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: the format from Task 1 and the tool from Task 3.
- Produces: user-facing documentation only.

- [ ] **Step 1: Write the failing check**

Run: `rg -n "Forbid:|critique" README.md`
Expected: no output (the README does not document the feature yet).

- [ ] **Step 2: Confirm the current README sections**

Run: `rg -n "^## |^\*\*\`rules.md\`\*\*|A rule without an" README.md`
Expected: shows the section headings, including `## Delivery tools`, and the `rules.md` paragraph about `Applies to:`.

- [ ] **Step 3: Write the documentation**

In `README.md`:

1. Immediately after the paragraph that begins "A rule without an `Applies to:` line is valid, but no delivery tool returns it", insert this new paragraph and example:

````markdown
A rule can also carry a check. The `Forbid:` line is a Go regular expression (RE2 syntax); each match on a line of a UI source file is a finding. The `Fix:` line says how to fix it. A rule has both lines or neither:

```markdown
## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
Forbid: #[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?)\(
Fix: Replace the raw colour with a colour token from tokens.json.
```

A pattern that does not compile, a repeated `Forbid:` or `Fix:`, or an empty one is a problem that `validate` reports. The rule is then not checked.
````

2. After the `## Delivery tools` section's last paragraph (the one ending "well under 4 KB."), add a new section:

````markdown
## Critique tool

`critique` checks one UI source file against the rules of a design system that have a `Forbid:` line. It returns one finding per match:

| Argument | Meaning |
|---|---|
| `file` | the source file, relative to the project root or absolute |
| `system` | the folder holding its `design-system/`, relative to the project root; `.` for the project root |

A finding reads:

```
testdata/critique/page.html:4: color.tokens-only: found "#ff0000". Replace the raw colour with a colour token from tokens.json.
```

The result names the file and the system, then one line per finding. When no rule can be checked, the result says so. The check is deterministic and reads the file as text, one line at a time, so a pattern cannot span lines and cannot judge layout or intent. The client that calls `critique` does that judgement, using the rules from the delivery tools. A `Forbid:` pattern is an approximation of its rule: the bundled spacing pattern flags every raw `px` value in `margin`, `padding` or `gap`, including values that are multiples of the unit. The source must be 1 MiB or less and a text file.
````

- [ ] **Step 4: Verify**

Run: `rg -n "Forbid:|## Critique tool" README.md`
Expected: the new lines appear.

Run: `go test ./...`
Expected: PASS (no code change; confirms nothing else moved).

**Model:** haiku (one file, exact text given, verification is grep).

---

### Task 5: ADR-0009 and the glossary term

**Files:**
- Create: `docs/adr/0009-rules-carry-deterministic-checks.md`
- Modify: `GLOSSARY.md`

**Interfaces:**
- Consumes: the decisions in the spec (C1, C4, C5, C20, C21).
- Produces: the record of the decision and the new term.

- [ ] **Step 1: Write the failing check**

Run: `test -f docs/adr/0009-rules-carry-deterministic-checks.md && echo present || echo missing`
Expected: `missing`.

- [ ] **Step 2: Write the ADR**

Create `docs/adr/0009-rules-carry-deterministic-checks.md`:

```markdown
---
status: accepted
---

# A rule can carry a deterministic pattern; judgement about a UI stays with the client

A rule may have a `Forbid:` line, a Go RE2 regular expression, and a `Fix:` line. The `critique` tool reports each match in a UI source file as a finding, naming the rule, the line, the matched text and the fix. Rules without a `Forbid:` line stay valid and are never checked by `critique`. The server has no model access, so it can only report what a pattern expresses. Judgement that needs the point of view or the user journey stays with the client, which reads the rules through the delivery tools and reasons over them. The trade-off is that a pattern is an approximation of its rule (the bundled spacing pattern flags some values that the rule allows), and the rule author must write and maintain the patterns. This keeps findings reproducible and testable.
```

- [ ] **Step 3: Add the glossary term**

In `GLOSSARY.md`, add this entry directly after the **Critique** entry (before **Done vs good**):

```markdown
**Finding**:
One breach of one rule in one file, with the line, the rule's id, the matched text and the rule's fix. Critique returns findings; it does not pass or fail a file.
_Avoid_: error, violation, issue, warning
```

- [ ] **Step 4: Verify**

Run: `test -f docs/adr/0009-rules-carry-deterministic-checks.md && rg -n "^\*\*Finding\*\*" GLOSSARY.md`
Expected: the ADR exists and the glossary line prints.

**Model:** haiku (two files, full content given, verification is a test command).

---

## Closing (the orchestrator, after all tasks)

- Run the full verification from B5: `go test ./...`, `gofmt -l .`, golangci-lint, `go build -o tenet ./cmd/tenet`, and `go run ./cmd/tenet --dir testdata/project` exits without printing to stdout.
- Tick `- [x] Slice 4: Critique tool` in `docs/roadmap.md`.
- One commit for the whole slice on `main`, with no co-author trailer.
