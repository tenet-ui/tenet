# Slice 6: Consistency experiment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Measure whether UI generated with the tenet MCP tools and the reference pack is more consistent than UI generated from a plain prompt, and commit a report that answers the question.

**Architecture:** An experiment folder, `experiments/consistency/`, with fixed briefs, a shell runner that calls `claude -p` ten times (five briefs, two conditions), raw HTML outputs, and a Go tally program that reuses `internal/designsystem` to count critique findings. No product code changes.

**Tech Stack:** Bash and awk (runner), Go 1.27 (tally, module `github.com/tenet-ui/tenet`), the `claude` CLI pinned to `claude-sonnet-5-5`.

**Spec:** `docs/superpowers/specs/2026-10-18-slice-6-consistency-experiment-design.md` (decision IDs E1–E22 below refer to it).

## Global Constraints

- No changes under `internal/` or `cmd/` (spec E1, E20).
- Every generation run uses `--model claude-sonnet-5-5` (E3) and `--max-budget-usd 2` (E10).
- Generation runs start in a `mktemp -d` directory outside the repository (E4).
- The tally and tests must pass `go test ./...`, `test -z "$(gofmt -l .)"`, and golangci-lint `v2.15.0` (E20).
- Do not run the tally against `experiments/consistency/outputs/` until all ten outputs (or `.failed` markers) exist (Task 7).

## Review Focus

1. A brief has both `<brief>.html` and `<brief>.failed` (a stale marker from an earlier failure) → the tally must refuse, not count both. Pinned by Task 2, `TestCollectRejectsStaleFailedFile`.
2. `tokens.json` holds an array, or its top level is not an object → an error, not a silently smaller token set. Pinned by Task 1, `TestFlattenTokensRejectsArrays` and `TestFlattenTokensRejectsNonObjectTop`.
3. A run exits 0 but returns no `<html` element, or exits non-zero → recorded as a failure with its own `.failed` file, never saved as an HTML output. Pinned by Task 4, the fake-claude `nohtml` and `fail` checks.
4. Brief text containing `$`, backticks or quotes reaches the model unchanged: the prompt goes to `claude` on stdin and is never passed through `eval` or a shell string. Pinned by Task 3, the `--print-prompt` checks and the no-`eval` check.
5. The outputs folder is missing or one condition has no outputs → the tally fails with a message; it never prints an empty table that looks like a result. Pinned by Task 2, `TestCollectMissingOutputsFails`.

---

### Task 1: Token helpers for the tally

**Files:**
- Create: `experiments/consistency/tally/tokens.go`
- Test: `experiments/consistency/tally/tokens_test.go`

**Interfaces:**
- Consumes: nothing (standard library only).
- Produces:
  - `func flattenTokens(data []byte) (map[string]bool, error)`: the set of CSS-style names for the leaf values of a `tokens.json` document. A leaf is any JSON value that is not an object. The name is `"--"` plus the key path joined with `"-"` (for example `color.ink` → `--color-ink`). An array value is an error. Invalid JSON, or a top-level value that is not an object, is an error.
  - `func unresolvedNames(source string, names map[string]bool) []string`: every name written `var(--name` in `source` that is not in `names`, in order of appearance, duplicates included. A name ends at the first character that is not a letter, digit, `-` or `_`.
  - `func hasVarRef(source string) bool`: true when `source` contains `var(--`.

- [ ] **Step 1: Write the failing tests**

Create `experiments/consistency/tally/tokens_test.go`:

```go
package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestFlattenTokensNamesLeaves(t *testing.T) {
	data := []byte(`{"color":{"ink":"#0d1b2e"},"space":{"md":16},"motion":{"standard":"cubic-bezier(0.25, 1, 0.5, 1)"}}`)
	got, err := flattenTokens(data)
	if err != nil {
		t.Fatalf("flattenTokens: %v", err)
	}
	want := []string{"--color-ink", "--motion-standard", "--space-md"}
	if names := sortedSet(got); !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
}

func TestFlattenTokensNestedPath(t *testing.T) {
	got, err := flattenTokens([]byte(`{"a":{"b":{"c":1}}}`))
	if err != nil {
		t.Fatalf("flattenTokens: %v", err)
	}
	if want := []string{"--a-b-c"}; !reflect.DeepEqual(sortedSet(got), want) {
		t.Fatalf("names = %v, want %v", sortedSet(got), want)
	}
}

func TestFlattenTokensRejectsArrays(t *testing.T) {
	if _, err := flattenTokens([]byte(`{"x":[1]}`)); err == nil {
		t.Fatal("expected an error for an array value")
	}
}

func TestFlattenTokensRejectsNonObjectTop(t *testing.T) {
	for _, data := range []string{`[1]`, `"text"`, `{"a":`} {
		if _, err := flattenTokens([]byte(data)); err == nil {
			t.Fatalf("expected an error for %s", data)
		}
	}
}

func TestUnresolvedNames(t *testing.T) {
	names := map[string]bool{"--color-ink": true}
	src := "color: var(--color-ink); margin: var(--space-zz); border: var(--space-zz, #fff)"
	want := []string{"--space-zz", "--space-zz"}
	if got := unresolvedNames(src, names); !reflect.DeepEqual(got, want) {
		t.Fatalf("unresolvedNames = %v, want %v", got, want)
	}
}

func TestUnresolvedNamesNoneWhenAllResolve(t *testing.T) {
	names := map[string]bool{"--color-ink": true}
	if got := unresolvedNames("color: var(--color-ink);", names); len(got) != 0 {
		t.Fatalf("unresolvedNames = %v, want none", got)
	}
}

func TestHasVarRef(t *testing.T) {
	cases := map[string]bool{
		"color: var(--color-ink);": true,
		"color: #fff;":             false,
		"var( --color-ink )":       false,
	}
	for src, want := range cases {
		if got := hasVarRef(src); got != want {
			t.Errorf("hasVarRef(%q) = %v, want %v", src, got, want)
		}
	}
}

func sortedSet(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./experiments/consistency/tally -run 'Flatten|Unresolved|HasVarRef' -v`
Expected: FAIL with `undefined: flattenTokens` (and the other two helpers).

- [ ] **Step 3: Write the implementation**

Create `experiments/consistency/tally/tokens.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var varRef = regexp.MustCompile(`var\(--([A-Za-z0-9_-]*)`)

// flattenTokens returns the CSS-style names of the leaf values in a tokens.json document.
func flattenTokens(data []byte) (map[string]bool, error) {
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("tokens.json: %w", err)
	}
	obj, ok := root.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("tokens.json: top level must be an object")
	}
	names := map[string]bool{}
	if err := collectNames(obj, nil, names); err != nil {
		return nil, err
	}
	return names, nil
}

func collectNames(obj map[string]any, path []string, names map[string]bool) error {
	for key, value := range obj {
		p := append(append([]string{}, path...), key)
		switch v := value.(type) {
		case map[string]any:
			if err := collectNames(v, p, names); err != nil {
				return err
			}
		case []any:
			return fmt.Errorf("tokens.json: %s is an array; tokens must be objects or values", strings.Join(p, "."))
		default:
			names["--"+strings.Join(p, "-")] = true
		}
	}
	return nil
}

// unresolvedNames returns each var(--name) in source whose name is not in names, in order.
func unresolvedNames(source string, names map[string]bool) []string {
	var out []string
	for _, m := range varRef.FindAllStringSubmatch(source, -1) {
		name := "--" + m[1]
		if !names[name] {
			out = append(out, name)
		}
	}
	return out
}

// hasVarRef reports whether source contains a var(-- reference.
func hasVarRef(source string) bool {
	return strings.Contains(source, "var(--")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./experiments/consistency/tally -v`
Expected: PASS for all six tests.

---

### Task 2: The tally program

**Files:**
- Create: `experiments/consistency/tally/main.go`
- Test: `experiments/consistency/tally/main_test.go`

**Interfaces:**
- Consumes (from Task 1): `flattenTokens`, `unresolvedNames`, `hasVarRef`.
- Consumes (from the product, read-only): `designsystem.Load(root string) (designsystem.DesignSystem, []designsystem.Problem)`, `designsystem.TokensPath` (`"design-system/tokens.json"`), `designsystem.Critique(source string, rules []designsystem.Rule) []designsystem.Finding`, `designsystem.DesignSystem.Rules`, `designsystem.Problem.String()`.
- Produces:
  - `type row struct { Condition, Brief, Status string; TokenOnly bool; Findings, Unresolved int }` where `Status` is `"ok"` or `"failed"`.
  - `func collect(packDir, outDir string) ([]row, error)`: rows for `plain` then `tools`, each sorted by brief ID. Errors: pack has problems; `tokens.json` unreadable or invalid; a condition folder missing; a brief with both `.html` and `.failed`; a brief missing from one condition.
  - `func render(rows []row) string`: the Markdown table.
  - `func summarize(rows []row) []string`: one line per condition.
  - `main` flags: `-pack` (default `packs/stripe-reference`), `-outputs` (default `experiments/consistency/outputs`). Prints the table, a blank line, then the summary lines. Exit 1 on error, message on stderr.
  - Token-only rule (E13): `Findings == 0 && hasVarRef(text)`. Unresolved count (E14): `len(unresolvedNames(text, names))`.
  - Summary line format (E15), with `N` = count of briefs in that condition: `- <cond>: <T> token-only of <N>; <F> findings in total; <M> findings per ok output; <OK> ok, <FAIL> failed`. `M` is one decimal, or `n/a` when there are no ok outputs. Failed outputs are excluded from `M`.

- [ ] **Step 1: Write the failing tests**

Create `experiments/consistency/tally/main_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPack = "../../../packs/stripe-reference"

func TestSummarizeCountsPerCondition(t *testing.T) {
	rows := []row{
		{Condition: "plain", Brief: "a", Status: "ok", TokenOnly: false, Findings: 4},
		{Condition: "plain", Brief: "b", Status: "failed"},
		{Condition: "tools", Brief: "a", Status: "ok", TokenOnly: true, Findings: 0},
		{Condition: "tools", Brief: "b", Status: "ok", TokenOnly: true, Findings: 1},
	}
	want := []string{
		"- plain: 0 token-only of 2; 4 findings in total; 4.0 findings per ok output; 1 ok, 1 failed",
		"- tools: 2 token-only of 2; 1 findings in total; 0.5 findings per ok output; 2 ok, 0 failed",
	}
	got := summarize(rows)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("summarize =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestRenderShowsFailedRowWithDashes(t *testing.T) {
	out := render([]row{{Condition: "plain", Brief: "sign-in", Status: "failed"}})
	if !strings.Contains(out, "| plain | sign-in | failed | no | – | – |") {
		t.Fatalf("render output missing failed row:\n%s", out)
	}
}

func TestCollectRejectsStaleFailedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "plain", "payments-list.html"), "<html></html>")
	writeFile(t, filepath.Join(dir, "plain", "payments-list.failed"), "boom")
	writeFile(t, filepath.Join(dir, "tools", "payments-list.html"), "<html></html>")
	_, err := collect(testPack, dir)
	if err == nil || !strings.Contains(err.Error(), "both .html and .failed") {
		t.Fatalf("collect error = %v, want a both-files error", err)
	}
}

func TestCollectMissingOutputsFails(t *testing.T) {
	_, err := collect(testPack, filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("collect with a missing outputs folder returned no error")
	}
}

func TestCollectCountsTokenOnlyAndFindings(t *testing.T) {
	dir := t.TempDir()
	clean := "<html><style>p { color: var(--color-ink); }</style></html>"
	dirty := "<html><style>p { color: #123456; }</style></html>"
	writeFile(t, filepath.Join(dir, "plain", "sign-in.html"), dirty)
	writeFile(t, filepath.Join(dir, "tools", "sign-in.html"), clean)
	rows, err := collect(testPack, dir)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	plain, tools := rows[0], rows[1]
	if plain.TokenOnly || plain.Findings == 0 {
		t.Errorf("plain row = %+v, want findings and not token-only", plain)
	}
	if !tools.TokenOnly || tools.Findings != 0 || tools.Unresolved != 0 {
		t.Errorf("tools row = %+v, want token-only with no findings and no unresolved names", tools)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./experiments/consistency/tally -run 'Summarize|Render|Collect' -v`
Expected: FAIL with `undefined: row`, `undefined: summarize`, `undefined: render`, `undefined: collect`.

- [ ] **Step 3: Write the implementation**

Create `experiments/consistency/tally/main.go`:

```go
// Command tally checks the consistency-experiment outputs against the reference pack
// and prints Markdown tables. Run it from the repository root.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tenet-ui/tenet/internal/designsystem"
)

var conditions = []string{"plain", "tools"}

type row struct {
	Condition  string
	Brief      string
	Status     string // "ok" or "failed"
	TokenOnly  bool
	Findings   int
	Unresolved int
}

func main() {
	packDir := flag.String("pack", "packs/stripe-reference", "pack folder, relative to the repository root")
	outDir := flag.String("outputs", "experiments/consistency/outputs", "outputs folder, relative to the repository root")
	flag.Parse()
	rows, err := collect(*packDir, *outDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(render(rows))
	fmt.Println()
	for _, line := range summarize(rows) {
		fmt.Println(line)
	}
}

func collect(packDir, outDir string) ([]row, error) {
	ds, problems := designsystem.Load(packDir)
	if len(problems) > 0 {
		var lines []string
		for _, p := range problems {
			lines = append(lines, p.String())
		}
		return nil, fmt.Errorf("pack %s has problems:\n%s", packDir, strings.Join(lines, "\n"))
	}
	tokens, err := os.ReadFile(filepath.Join(packDir, designsystem.TokensPath))
	if err != nil {
		return nil, err
	}
	names, err := flattenTokens(tokens)
	if err != nil {
		return nil, err
	}

	briefs := map[string]bool{}
	for _, cond := range conditions {
		entries, err := os.ReadDir(filepath.Join(outDir, cond))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if id, ok := briefID(e.Name()); ok {
				briefs[id] = true
			}
		}
	}

	var rows []row
	for _, cond := range conditions {
		for _, id := range sortedKeys(briefs) {
			r, err := readRow(ds.Rules, names, filepath.Join(outDir, cond), cond, id)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r)
		}
	}
	return rows, nil
}

func readRow(rules []designsystem.Rule, names map[string]bool, dir, cond, id string) (row, error) {
	htmlPath := filepath.Join(dir, id+".html")
	_, htmlErr := os.Stat(htmlPath)
	_, failedErr := os.Stat(filepath.Join(dir, id+".failed"))
	hasHTML, hasFailed := htmlErr == nil, failedErr == nil
	switch {
	case hasHTML && hasFailed:
		return row{}, fmt.Errorf("%s/%s: both .html and .failed exist; remove the stale one", cond, id)
	case hasFailed:
		return row{Condition: cond, Brief: id, Status: "failed"}, nil
	case hasHTML:
		src, err := os.ReadFile(htmlPath)
		if err != nil {
			return row{}, err
		}
		text := string(src)
		findings := len(designsystem.Critique(text, rules))
		return row{
			Condition:  cond,
			Brief:      id,
			Status:     "ok",
			TokenOnly:  findings == 0 && hasVarRef(text),
			Findings:   findings,
			Unresolved: len(unresolvedNames(text, names)),
		}, nil
	default:
		return row{}, fmt.Errorf("%s/%s: no .html or .failed file", cond, id)
	}
}

func render(rows []row) string {
	var b strings.Builder
	b.WriteString("| Condition | Brief | Status | Token-only | Critique findings | Unresolved var names |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, r := range rows {
		if r.Status == "failed" {
			fmt.Fprintf(&b, "| %s | %s | failed | no | – | – |\n", r.Condition, r.Brief)
			continue
		}
		fmt.Fprintf(&b, "| %s | %s | ok | %s | %d | %d |\n", r.Condition, r.Brief, yesNo(r.TokenOnly), r.Findings, r.Unresolved)
	}
	return b.String()
}

func summarize(rows []row) []string {
	var lines []string
	for _, cond := range conditions {
		var ok, failed, tokenOnly, findings int
		for _, r := range rows {
			if r.Condition != cond {
				continue
			}
			if r.Status == "failed" {
				failed++
				continue
			}
			ok++
			findings += r.Findings
			if r.TokenOnly {
				tokenOnly++
			}
		}
		mean := "n/a"
		if ok > 0 {
			mean = fmt.Sprintf("%.1f", float64(findings)/float64(ok))
		}
		lines = append(lines, fmt.Sprintf("- %s: %d token-only of %d; %d findings in total; %s findings per ok output; %d ok, %d failed",
			cond, tokenOnly, ok+failed, findings, mean, ok, failed))
	}
	return lines
}

func briefID(name string) (string, bool) {
	for _, ext := range []string{".html", ".failed"} {
		if strings.HasSuffix(name, ext) {
			return strings.TrimSuffix(name, ext), true
		}
	}
	return "", false
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./experiments/consistency/tally -v`
Expected: PASS for all tests in Tasks 1 and 2.

Run: `go vet ./experiments/consistency/tally`
Expected: no output.

**Model: sonnet**

---

### Task 3: Briefs and runner prompt building

**Files:**
- Create: `experiments/consistency/briefs.md`
- Create: `experiments/consistency/run.sh`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `briefs.md` sections, found by exact heading: `## contract`, `## addendum-tools`, and one `## brief: <id>` per brief, in this order: `payments-list`, `empty-state`, `account-setup`, `sign-in`, `settings-card`.
  - `run.sh --print-prompt <plain|tools> <brief-id>`: prints the exact prompt for that run. The tools prompt is the brief, then the addendum, then the contract. The plain prompt is the brief, then the contract.
  - `run.sh --dry-run [--only <brief-id>]`: prints one line `would run: <cond> <id>` per selected run (plain before tools, briefs in file order) and exits 0 without calling `claude`.
  - Variables: `MODEL=claude-sonnet-5-5`, `BARE_FLAGS=(--bare)`, `out_dir` from `OUT_DIR` or `$here/outputs`.

- [ ] **Step 1: Write the failing checks**

Run: `bash experiments/consistency/run.sh --dry-run | wc -l`
Expected: FAIL with `No such file or directory` (the runner does not exist yet).

- [ ] **Step 2: Create `experiments/consistency/briefs.md`**

Create the file with exactly this content:

````markdown
# Consistency experiment: briefs and prompt parts

Each `## brief: <id>` section is one fixed brief. The runner builds each prompt from one brief, the `addendum-tools` section (tools condition only), and the `contract` section. Do not edit these texts after the first run; the report depends on them.

## contract

Reply with one complete HTML document and nothing else. Put all CSS in one `<style>` element in the head. Use no external assets, no external fonts, no images from URLs, and no JavaScript. Do not add commentary before or after the document.

## addendum-tools

The project's design system is available through the tenet MCP tools. Before you write the page, call `get_point_of_view`, then call `get_tokens`, `get_component`, `get_template` or `get_flow` for each item that fits the brief. Build the page only from the tokens those tools return.

## brief: payments-list

Build a payments list page for a small online business. Each row shows the amount, the customer, the status and the date, and opens the payment's detail page. Include a filter control and an export button. The page has a heading and a short description under it.

## brief: empty-state

Build the empty state shown on a payments list when the business has no payments yet. It has a title, one line of supporting text, and one action that creates a payment link.

## brief: account-setup

Build the first step of account setup for a small online business. Ask for the business name and the country, and give the user a continue action and a back action.

## brief: sign-in

Build a sign-in form with an email field and a password field, a primary sign-in button, and a secondary action to create an account.

## brief: settings-card

Build a settings card with a title, a one-sentence description, and three actions: save, cancel, and delete the account.
````

- [ ] **Step 3: Create `experiments/consistency/run.sh`**

Create the file with exactly this content, then make it executable with `chmod +x experiments/consistency/run.sh`:

```bash
#!/usr/bin/env bash
# Consistency experiment runner (slice 6). Prompts and conditions: briefs.md and the spec.
set -euo pipefail

here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../.." && pwd)
briefs="$here/briefs.md"
out_dir=${OUT_DIR:-$here/outputs}
MODEL=claude-sonnet-5-5
# --bare skips keychain reads. If the smoke check in REPORT.md shows auth fails with it, set this to ().
BARE_FLAGS=(--bare)

# section <heading>: prints the body of "## <heading>" up to the next "## " heading.
section() {
  awk -v h="## $1" '$0 == h {on = 1; next} /^## / {on = 0} on' "$briefs" | sed '/./,$!d'
}

brief_ids() {
  sed -n 's/^## brief: //p' "$briefs"
}

# build_prompt <plain|tools> <brief-id>
build_prompt() {
  section "brief: $2"
  echo
  if [ "$1" = tools ]; then
    section addendum-tools
    echo
  fi
  section contract
}

if [ "${1:-}" = --print-prompt ]; then
  build_prompt "$2" "$3"
  exit 0
fi

dry=0
force=0
only=""
while [ $# -gt 0 ]; do
  case $1 in
    --dry-run) dry=1 ;;
    --force) force=1 ;;
    --only) only=$2; shift ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
  shift
done

selected() { [ -z "$only" ] || [ "$only" = "$1" ]; }

if [ "$dry" = 1 ]; then
  for cond in plain tools; do
    for id in $(brief_ids); do
      if selected "$id"; then
        echo "would run: $cond $id"
      fi
    done
  done
  exit 0
fi
```

- [ ] **Step 4: Run the checks to verify they pass**

Run: `bash experiments/consistency/run.sh --dry-run | wc -l`
Expected: `10`

Run: `bash experiments/consistency/run.sh --dry-run --only sign-in`
Expected: two lines, `would run: plain sign-in` and `would run: tools sign-in`.

Run: `bash experiments/consistency/run.sh --print-prompt plain sign-in | grep -c 'get_point_of_view'`
Expected: `0` (grep prints 0 and exits 1; that exit is expected here).

Run: `bash experiments/consistency/run.sh --print-prompt tools sign-in | grep -c 'get_point_of_view'`
Expected: `1`

Run: `bash experiments/consistency/run.sh --print-prompt tools sign-in | grep -c 'HTML document'`
Expected: `1`

Run: `grep -c 'eval' experiments/consistency/run.sh`
Expected: `0` (grep exits 1 with count 0; that is the check).

**Model: sonnet**

---

### Task 4: Runner live path

**Files:**
- Modify: `experiments/consistency/run.sh` (append the live section from this task to the end of the file)
- Test: a throwaway fake `claude` in a `mktemp -d` directory, not committed

**Interfaces:**
- Consumes (from Task 3): `build_prompt`, `brief_ids`, `selected`, `out_dir`, `MODEL`, `BARE_FLAGS`, `repo`, `here`, `force`, `only`.
- Produces: live runs. For each selected run, the output is written to `$out_dir/<cond>/<id>.html` on success or `$out_dir/<cond>/<id>.failed` (the error text) on failure. Each run gets two attempts (E9). An existing `.html` is skipped unless `--force` (so reruns cost nothing for finished runs). Exit status is 0 only when no run failed.
- Live command, plain: `claude -p "${BARE_FLAGS[@]}" --strict-mcp-config --mcp-config <none.json> --model claude-sonnet-5-5 --no-session-persistence --max-budget-usd 2 --tools ""`, with the prompt on stdin, run from the temp directory.
- Live command, tools: the same, with `--mcp-config <mcp.json> --allowedTools "mcp__tenet__*"` in place of the empty config.
- Code fences are removed from the reply (`sed '/^```/d'`). A reply with no `<html` (case-insensitive) is a failed attempt.

- [ ] **Step 1: Write the failing checks with a fake `claude`**

Run: `OUT_DIR=$(mktemp -d) bash experiments/consistency/run.sh --only sign-in; ls "$OUT_DIR"`
Expected: FAIL, because the live path does not exist yet. The listing is empty, so no `plain` or `tools` folder was written.

- [ ] **Step 2: Append the live section to `experiments/consistency/run.sh`**

Append exactly this to the end of the file:

```bash

if ! command -v claude >/dev/null 2>&1; then
  echo "claude CLI not found on PATH" >&2
  exit 1
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
(cd "$repo" && go build -o "$work/tenet" ./cmd/tenet)
cat > "$work/mcp.json" <<EOF
{"mcpServers":{"tenet":{"command":"$work/tenet","args":["--dir","$repo/packs/stripe-reference"]}}}
EOF
echo '{"mcpServers":{}}' > "$work/none.json"
cd "$work"

# run_one <plain|tools> <brief-id>
run_one() {
  local cond=$1 id=$2 dir="$out_dir/$1" attempt err out prompt
  local cfg=()
  mkdir -p "$dir"
  if [ "$force" = 0 ] && [ -f "$dir/$id.html" ]; then
    echo "skip $cond $id (already has output)" >&2
    return 0
  fi
  if [ "$cond" = tools ]; then
    cfg=(--mcp-config "$work/mcp.json" --allowedTools "mcp__tenet__*")
  else
    cfg=(--mcp-config "$work/none.json")
  fi
  prompt=$(build_prompt "$cond" "$id")
  err=""
  for attempt in 1 2; do
    err=""
    if out=$(printf '%s' "$prompt" | claude -p "${BARE_FLAGS[@]}" --strict-mcp-config "${cfg[@]}" --model "$MODEL" --no-session-persistence --max-budget-usd 2 --tools "" 2>"$work/err.txt"); then
      printf '%s\n' "$out" | sed '/^```/d' > "$work/candidate.html"
      if grep -qi '<html' "$work/candidate.html"; then
        break
      fi
      err="reply had no <html> element"
    else
      err=$(cat "$work/err.txt")
    fi
  done
  if [ -z "$err" ]; then
    mv "$work/candidate.html" "$dir/$id.html"
    rm -f "$dir/$id.failed"
    echo "ok $cond $id" >&2
  else
    rm -f "$dir/$id.html"
    printf '%s\n' "$err" > "$dir/$id.failed"
    echo "FAILED $cond $id: $err" >&2
    return 1
  fi
}

failures=0
for cond in plain tools; do
  for id in $(brief_ids); do
    if selected "$id"; then
      run_one "$cond" "$id" || failures=$((failures + 1))
    fi
  done
done
echo "done: $failures failed" >&2
[ "$failures" -eq 0 ]
```

- [ ] **Step 3: Run the fake-claude checks to verify they pass**

Create the fake CLI once, outside the repository:

```bash
fake=$(mktemp -d)
mkdir -p "$fake/bin"
cat > "$fake/bin/claude" <<'EOF'
#!/usr/bin/env bash
cat >/dev/null
case "${FAKE_CLAUDE:-html}" in
  html) printf '```html\n<!doctype html><html><body><p>ok</p></body></html>\n```\n' ;;
  nohtml) echo "sorry, no page" ;;
  fail) echo "boom" >&2; exit 1 ;;
esac
EOF
chmod +x "$fake/bin/claude"
```

Check 1, success with fences stripped:
`PATH="$fake/bin:$PATH" OUT_DIR="$fake/out" bash experiments/consistency/run.sh --only sign-in`
Expected: exit 0; `ls "$fake/out/plain" "$fake/out/tools"` shows `sign-in.html` in both; `grep -c '```' "$fake/out/plain/sign-in.html"` prints `0`.

Check 2, rerun skips finished output:
`PATH="$fake/bin:$PATH" OUT_DIR="$fake/out" bash experiments/consistency/run.sh --only sign-in`
Expected: exit 0; stderr has `skip plain sign-in` and `skip tools sign-in`.

Check 3, exit non-zero is a failure, not a saved output:
`FAKE_CLAUDE=fail PATH="$fake/bin:$PATH" OUT_DIR="$fake/out2" bash experiments/consistency/run.sh --only empty-state`
Expected: exit 1; `$fake/out2/plain/empty-state.failed` contains `boom`; no `empty-state.html` in either condition folder.

Check 4, a reply with no `<html>` is a failure:
`FAKE_CLAUDE=nohtml PATH="$fake/bin:$PATH" OUT_DIR="$fake/out3" bash experiments/consistency/run.sh --only account-setup`
Expected: exit 1; `$fake/out3/tools/account-setup.failed` contains `reply had no <html> element`.

Check 5, missing CLI:
`env PATH=/nonexistent /bin/bash experiments/consistency/run.sh --only sign-in`
Expected: exit 1, stderr `claude CLI not found on PATH`.

Check 6, only the tools condition gets MCP tools: `grep -n 'allowedTools' experiments/consistency/run.sh`. Expected: one match, inside the `tools` branch of `run_one`.

Remove the fake outputs with `rm -rf "$fake"` once the checks pass.

**Model: sonnet**

---

### Task 5: Smoke checks and the first two runs

This task is run by the session, not dispatched. It is the spec's E22 check: the first plain and first tools run are inspected by hand before the other eight start. Runs cost money; each is capped at 2 USD.

**Files:**
- Modify (only if a smoke check forces it): `experiments/consistency/run.sh` (`BARE_FLAGS`, or `--tools ""`).
- Outputs: `experiments/consistency/outputs/plain/payments-list.html`, `experiments/consistency/outputs/tools/payments-list.html`.

- [ ] **Step 1: Build the tenet binary and the MCP config in a scratch directory**

```bash
scratch=/tmp/claude-1000/-home-hossein-Projects-github-com-tenet-ui-tenet/8268f432-3b21-4be5-8a40-21f12fba76a8/scratchpad/slice6
mkdir -p "$scratch"
cd /home/hossein/Projects/github.com/tenet-ui/tenet && go build -o "$scratch/tenet" ./cmd/tenet
cat > "$scratch/mcp.json" <<EOF
{"mcpServers":{"tenet":{"command":"$scratch/tenet","args":["--dir","/home/hossein/Projects/github.com/tenet-ui/tenet/packs/stripe-reference"]}}}
EOF
echo '{"mcpServers":{}}' > "$scratch/none.json"
```

Expected: the build succeeds; both JSON files exist.

- [ ] **Step 2: Smoke A, does `--bare` authenticate**

```bash
cd "$scratch" && claude -p --bare --model claude-sonnet-5-5 --no-session-persistence --tools "" <<< "Reply with the word OK and nothing else."
```

Expected: `OK`. If it reports an authentication or login error, the keychain is needed: change `BARE_FLAGS=(--bare)` in `experiments/consistency/run.sh` to `BARE_FLAGS=()`, and record in REPORT.md's limits that `--bare` was dropped and why. Then repeat this step without `--bare` and expect `OK`.

- [ ] **Step 3: Smoke B, the tools condition reaches the server**

```bash
cd "$scratch" && claude -p --bare --strict-mcp-config --mcp-config "$scratch/mcp.json" --allowedTools "mcp__tenet__*" --model claude-sonnet-5-5 --no-session-persistence --tools "" <<< "Call the get_point_of_view tool, then reply with only the first line of its result."
head -1 /home/hossein/Projects/github.com/tenet-ui/tenet/packs/stripe-reference/design-system/point-of-view.md
```

Drop `--bare` from this command if Step 2 showed it must be dropped (the `BARE_FLAGS` value in `run.sh` is then `()`). Expected: the reply equals the `head -1` output. If the reply says it has no such tool, `--tools ""` is blocking MCP tools: remove `--tools ""` from this command and from the tools branch of `run.sh`, and record that in REPORT.md's limits.

- [ ] **Step 4: Smoke C, the plain condition has no tools**

```bash
cd "$scratch" && claude -p --strict-mcp-config --mcp-config "$scratch/none.json" --model claude-sonnet-5-5 --no-session-persistence --tools "" <<< "List any tools you can call. If you have none, reply NONE."
```

Expected: `NONE`.

- [ ] **Step 5: The first two runs, then inspect them**

```bash
cd /home/hossein/Projects/github.com/tenet-ui/tenet && bash experiments/consistency/run.sh --only payments-list
```

Expected: `ok plain payments-list` and `ok tools payments-list` on stderr, exit 0. Check each file:

```bash
for f in experiments/consistency/outputs/plain/payments-list.html experiments/consistency/outputs/tools/payments-list.html; do
  echo "== $f"; head -c 200 "$f"; echo; grep -c '```' "$f"; grep -c 'var(--' "$f"; grep -ci 'tool' "$f"
done
```

Expected: each file starts with `<!doctype html>` or `<html`; the fence count is `0`; the tools file has at least one `var(--` (the plain file may have none). Read both files in full. If a file contains commentary outside the document, or the tools reply is not a page built from pack tokens, record that in the report's limits. Do not edit the outputs.

Do not start Task 6 until both files are inspected and the smoke checks pass.

**Model: sonnet**

---

### Task 6: Remaining eight runs

**Files:**
- Outputs: the eight remaining `experiments/consistency/outputs/<cond>/<id>.html` or `.failed` files.

- [ ] **Step 1: Run every remaining brief in both conditions**

Run: `cd /home/hossein/Projects/github.com/tenet-ui/tenet && bash experiments/consistency/run.sh`
Expected: `skip plain payments-list` and `skip tools payments-list` on stderr, then `ok` or `FAILED` lines for the other eight runs, then `done: <n> failed`. The exit status is 0 when every run succeeded.

- [ ] **Step 2: Check the file count**

Run: `ls experiments/consistency/outputs/plain experiments/consistency/outputs/tools`
Expected: each folder has five entries, each named `<id>.html` or `<id>.failed`, with ids `account-setup`, `empty-state`, `payments-list`, `settings-card`, `sign-in`.

- [ ] **Step 3: Inspect each new output**

For each `.html`, check the first bytes and the fence count as in Task 5, Step 5. For each `.failed`, read its text. Do not rerun a failed run manually; it is recorded as a failure (E9). Record every failure for the report.

**Model: sonnet**

---

### Task 7: Tally and the report

**Files:**
- Create: `experiments/consistency/REPORT.md`

**Interfaces:**
- Consumes: the tally output (`go run ./experiments/consistency/tally`) and the spec's E16 and E17 rules.

- [ ] **Step 1: Run the tally**

Run: `cd /home/hossein/Projects/github.com/tenet-ui/tenet && go run ./experiments/consistency/tally`
Expected: exit 0; a Markdown table with ten rows and two summary lines. Save its output to the scratchpad file `/tmp/claude-1000/-home-hossein-Projects-github-com-tenet-ui-tenet/8268f432-3b21-4be5-8a40-21f12fba76a8/scratchpad/slice6/tally.md`. If it fails, fix the cause (a stale `.failed`, or a missing outputs folder) and rerun; do not hand-edit the numbers.

- [ ] **Step 2: Decide the answer by the E17 rule**

Let `P` be the plain token-only count and `T` the tools token-only count; let `PF` and `TF` be the total findings for each condition from the summary lines.
- "Improved": `T > P` AND `TF < PF`.
- "Mixed": exactly one of `T > P` or `TF < PF` holds.
- "Did not improve": neither holds.

Write the rule's result in the report. Do not change the rule after seeing the numbers.

- [ ] **Step 3: Write `experiments/consistency/REPORT.md`**

The report has these sections, in this order:
1. **Question**: does generating UI with the tenet tools and the reference pack make it more consistent than a plain prompt? (Roadmap Slice 6; ADR-0003's risk.)
2. **Setup**: date 2026-10-18; model `claude-sonnet-5-5`; output of `claude --version` (run it and paste); five briefs in `briefs.md`, one sample per brief per condition; the pack `packs/stripe-reference`; the two conditions as in spec E5 and E6; the metric definitions from spec E12–E14.
3. **Results**: the tally table pasted verbatim from the scratchpad `tally.md`, then the two summary lines verbatim.
4. **Failures**: each `.failed` output with its text, or "None".
5. **Answer**: one sentence, the E17 result ("Improved", "Mixed", or "Did not improve") followed by the numbers that decide it. Then one sentence on what it means for ADR-0003.
6. **Limits**: five briefs and one sample per condition, so this is direction, not proof (E16); one model (E3); critique checks text against `Forbid:` patterns only, so judgement about the point of view and the user journey is not measured; token-only is a proxy for the presence of named values, not their correct use; unresolved names depend on the flattened naming convention (E14); the `--bare` or `--tools ""` decisions from Task 5, if any were changed.

Note: the report states the numbers and the answer. It does not argue for the result beyond the rule.

**Model: sonnet**

---

### Task 8: ADR-0003 result

**Files:**
- Modify: `docs/adr/0003-on-demand-delivery-through-tools.md` (append a `## Result` section)

**Interfaces:**
- Consumes: the Answer sentence from `experiments/consistency/REPORT.md` (Task 7, Step 3, section 5).

- [ ] **Step 1: Append the Result section**

Append to `docs/adr/0003-on-demand-delivery-through-tools.md`:

```markdown

## Result (2026-10-18)

The consistency experiment (slice 6, [experiments/consistency/REPORT.md](../../experiments/consistency/REPORT.md)) ran five briefs in both conditions. Answer: <the Answer sentence from REPORT.md, copied verbatim>.
```

- [ ] **Step 2: Check the decision**

Re-read the ADR's decision text. Do not change it. If the answer is "Did not improve", the decision is contradicted by the report: add the sentence "This decision is under review; see the supersession rule in the slice flow." and stop, and flag it in the final summary so the human can decide whether a new ADR supersedes this one. This is a judgement the spec leaves to the report (E18); record the call in the final summary.

**Model: haiku**

---

### Task 9: Verify and tick the slice

**Files:**
- Modify: `docs/roadmap.md` (tick Slice 6: `- [ ] Slice 6` becomes `- [x] Slice 6`)

- [ ] **Step 1: Full suite, format, lint, build, tally**

```bash
cd /home/hossein/Projects/github.com/tenet-ui/tenet
go test ./...
test -z "$(gofmt -l .)" && echo "gofmt clean"
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
go build -o tenet ./cmd/tenet
go run ./experiments/consistency/tally > /dev/null && echo "tally ok"
```

Expected: `go test` passes; `gofmt clean`; golangci-lint prints no issues; the build succeeds; `tally ok`. If any step fails, stop and report the output; do not tick the roadmap.

- [ ] **Step 2: Check each acceptance criterion against the files**

- Criterion 1, five briefs each run twice: `ls experiments/consistency/outputs/plain experiments/consistency/outputs/tools` shows five entries each.
- Criterion 2, committed report: `test -f experiments/consistency/REPORT.md` succeeds, and the report's table has ten rows.

- [ ] **Step 3: Tick the roadmap**

In `docs/roadmap.md`, change the line `- [ ] Slice 6: Consistency experiment` to `- [x] Slice 6: Consistency experiment`.

**Model: haiku**
