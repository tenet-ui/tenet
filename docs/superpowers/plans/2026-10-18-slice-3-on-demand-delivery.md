# Slice 3: On-demand delivery tools Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose four MCP tools (`get_tokens`, `get_component`, `get_template`, `get_flow`) that return one requested item plus the valid rules linked to it, with errors that name what is missing and list what exists.

**Architecture:** Rules gain an optional `Applies to:` line that links them to items and token categories; `designsystem` gains `Ref`, `Rule.AppliesTo`, `Rule.Valid`, `Item.Body`, and pure lookup methods on `DesignSystem`. `Load` checks the links. The server registers four tools; each call loads the pack fresh and renders the result through one function that owns the response shape.

**Tech Stack:** Go 1.27 (`go.mod`), `github.com/modelcontextprotocol/go-sdk` v1.8.0, standard library only in `internal/designsystem`.

**Spec:** `docs/superpowers/specs/2026-10-18-slice-3-on-demand-delivery-design.md` (approved, commit 2dc2d84). Decision numbers below (D1–D19) refer to that spec.

## Global Constraints

- Module `github.com/tenet-ui/tenet`; Go 1.27; go-sdk `v1.8.0`. No new dependencies.
- `internal/designsystem` imports only the standard library and has no MCP types.
- While serving, stdout carries only MCP protocol; logs go to stderr.
- Tool failures are `IsError` tool results, not Go errors (as in slice 1).
- The bundled example's delivery result for any one call is under 4096 bytes.
- Do not add commit steps. Implementers and fixers do not commit; the orchestrator makes one commit at the end.
- Test command for everything: `go test ./...`; format: `test -z "$(gofmt -l .)"`; lint: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`.

## Review Focus

1. Empty `id` or `category` argument → not-found error that lists what exists, not a crash or a bare error. Test: Task 7 `TestEmptyIDListsAvailable`.
2. A rule linked to the same target twice → the rule appears once in the response. Test: Task 3 `TestLookups` (rule `a` in the sample).
3. An item file with a UTF-8 BOM → `Body` and `Title` have no BOM. Test: Task 1 `TestLoadItemsBody`.
4. `Applies to:` with spaces around entries is accepted; a trailing comma is a problem. Tests: Task 2 cases `spaces around entries` and `trailing comma`.
5. A numeric token value renders as its literal text (`space.unit: 4`). Test: Task 7 `TestGetTokensNumberValue`.

## Module shape

The four lookups on `DesignSystem` (`Category`, `Item`, `ItemIDs`, `Linked`) are the seam. Their small interface hides the filtering rules: only valid rules, kind and id must both match, rules keep `rules.md` order. Without `Linked`, that filter would reappear in each of the four handlers. `deliveryText` in `internal/server` is the one place that defines the response shape, so the four tools cannot drift apart in format.

---

### Task 1: Item body and rule link types

**Files:**
- Modify: `internal/designsystem/load.go` (add `Kind`, `Ref`; add `Rule.AppliesTo`, `Rule.Valid`; add `Item.Body`)
- Modify: `internal/designsystem/items.go` (set `Body` in `loadItems`)
- Test: `internal/designsystem/items_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: `type Kind string` with `KindComponent = "component"`, `KindTemplate = "template"`, `KindFlow = "flow"`, `KindTokens = "tokens"`; `type Ref struct { Kind Kind; ID string; Line int }`; `Rule.AppliesTo []Ref`; `Rule.Valid bool`; `Item.Body string` (full file text, BOM stripped).

**Model:** sonnet

- [ ] **Step 1: Write the failing test**

Append to `internal/designsystem/items_test.go`:

```go
func TestLoadItemsBody(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"design-system/components/button.md": "\xEF\xBB\xBF# Button\n\nAn action.\n",
	})
	items, problems := loadItems(root, ComponentsDir)
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].Title != "Button" {
		t.Errorf("Title = %q, want %q", items[0].Title, "Button")
	}
	if want := "# Button\n\nAn action.\n"; items[0].Body != want {
		t.Errorf("Body = %q, want %q", items[0].Body, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/designsystem -run TestLoadItemsBody -v`
Expected: FAIL to compile with `items[0].Body undefined`.

- [ ] **Step 3: Write minimal implementation**

In `internal/designsystem/load.go`, add above `type Rule struct`:

```go
// Kind names what a rule can apply to. Its values are the kind part of a "kind:id" ref.
type Kind string

const (
	KindComponent Kind = "component"
	KindTemplate  Kind = "template"
	KindFlow      Kind = "flow"
	KindTokens    Kind = "tokens" // id is a token category name
)

// Ref is one entry of a rule's Applies to: line.
type Ref struct {
	Kind Kind
	ID   string
	Line int // 1-based line of the Applies to: line in rules.md
}
```

Change `Rule` to:

```go
// Rule is one "## <id>" entry of rules.md.
type Rule struct {
	ID        string
	Statement string // body lines other than the Reason: line, Applies to: line and fenced code, joined with a space
	Reason    string
	Line      int   // line of the "## <id>" heading
	AppliesTo []Ref // refs in the order written
	Valid     bool  // false when parseRules reported any problem for this entry
}
```

Change `Item` to add the field:

```go
// Item is one Markdown file in components/, templates/ or flows/.
type Item struct {
	ID    string // file name without ".md"
	Title string // text of the first "# " heading
	Path  string // slash-separated, relative to the project root
	Body  string // full file text, BOM stripped
}
```

In `internal/designsystem/items.go`, change the append line in `loadItems` to:

```go
		items = append(items, Item{ID: id, Title: title, Path: file, Body: string(stripBOM(data))})
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/designsystem -v`
Expected: PASS for all tests, including `TestLoadItemsBody` and the existing `TestLoadItems`.

---

### Task 2: Parse the `Applies to:` line and the `Valid` flag

**Files:**
- Modify: `internal/designsystem/rules.go`
- Test: `internal/designsystem/rules_test.go`

**Interfaces:**
- Consumes: `Kind`, `Ref`, `Rule.AppliesTo`, `Rule.Valid` from Task 1; `itemIDPattern` (items.go); `namePattern` (tokens.go).
- Produces: `parseRef(ruleID, entry string, line int) (Ref, string)` returning the ref, or a problem message when invalid; `parseRules` now fills `AppliesTo` and `Valid`.

**Model:** sonnet

- [ ] **Step 1: Write the failing test**

Append to `internal/designsystem/rules_test.go`:

```go
// formatRefs renders refs as "kind:id@line,..." for comparison.
func formatRefs(rs []Ref) string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, fmt.Sprintf("%s:%s@%d", r.Kind, r.ID, r.Line))
	}
	return strings.Join(out, ",")
}

func TestParseRulesAppliesTo(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantRules    []string // "id|refs|valid"
		wantProblems []wantProblem
	}{
		{
			name: "two refs, valid",
			data: joinLines(
				"# Rules",
				"",
				"## color.tokens-only",
				"Components take colours from tokens.",
				"Reason: Raw values drift.",
				"Applies to: tokens:color, component:button",
			),
			wantRules: []string{"color.tokens-only|tokens:color@6,component:button@6|true"},
		},
		{
			name: "spaces around entries are accepted",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to:  tokens:color ,component:button  ",
			),
			wantRules: []string{"a.b|tokens:color@4,component:button@4|true"},
		},
		{
			name: "applies to line is not part of the statement",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:button",
			),
			wantRules: []string{"a.b|component:button@4|true"},
		},
		{
			name: "unknown kind is reported and the valid ref is kept",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: thing:x, component:button",
			),
			wantRules: []string{"a.b|component:button@4|false"},
			wantProblems: []wantProblem{
				{4, `unknown kind "thing" in Applies to:`},
			},
		},
		{
			name: "entry without a colon",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: button",
			),
			wantRules: []string{"a.b||false"},
			wantProblems: []wantProblem{
				{4, `without a kind; write kind:id`},
			},
		},
		{
			name: "invalid item id",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:Button",
			),
			wantRules: []string{"a.b||false"},
			wantProblems: []wantProblem{
				{4, `invalid id "Button" in Applies to:`},
			},
		},
		{
			name: "trailing comma is an empty entry",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:button,",
			),
			wantRules: []string{"a.b|component:button@4|false"},
			wantProblems: []wantProblem{
				{4, "empty entry in Applies to:"},
			},
		},
		{
			name: "empty applies to line",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to:",
			),
			wantRules: []string{"a.b||false"},
			wantProblems: []wantProblem{
				{4, "empty Applies to: line"},
			},
		},
		{
			name: "second applies to line is reported and the first stands",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:button",
				"Applies to: template:settings-page",
			),
			wantRules: []string{"a.b|component:button@4|false"},
			wantProblems: []wantProblem{
				{5, "more than one Applies to: line"},
			},
		},
		{
			name: "applies to inside a fence is ignored",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"```",
				"Applies to: component:x",
				"```",
			),
			wantRules: []string{"a.b||true"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, problems := parseRules("design-system/rules.md", []byte(tt.data))
			got := make([]string, 0, len(rules))
			for _, r := range rules {
				got = append(got, fmt.Sprintf("%s|%s|%t", r.ID, formatRefs(r.AppliesTo), r.Valid))
			}
			if !reflect.DeepEqual(got, tt.wantRules) {
				t.Errorf("rules = %v, want %v", got, tt.wantRules)
			}
			checkProblems(t, problems, tt.wantProblems)
			for _, r := range rules {
				if strings.Contains(r.Statement, "Applies to") {
					t.Errorf("statement %q contains the Applies to: line", r.Statement)
				}
			}
		})
	}
}
```

Add `"reflect"` and `"strings"` to the imports of `rules_test.go` if they are not already there.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/designsystem -run TestParseRulesAppliesTo -v`
Expected: FAIL: the `wantRules` strings do not match (refs are empty and `Valid` is always true).

- [ ] **Step 3: Write minimal implementation**

In `internal/designsystem/rules.go`, add this helper below `ruleIDPattern`:

```go
// parseRef parses one entry of an Applies to: line. It returns the message to report when the entry is invalid.
func parseRef(ruleID, entry string, line int) (Ref, string) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return Ref{}, fmt.Sprintf("rule %q has an empty entry in Applies to:", ruleID)
	}
	kindText, id, found := strings.Cut(entry, ":")
	if !found {
		return Ref{}, fmt.Sprintf("rule %q has an Applies to: entry %q without a kind; write kind:id", ruleID, entry)
	}
	kind := Kind(strings.TrimSpace(kindText))
	id = strings.TrimSpace(id)
	switch kind {
	case KindComponent, KindTemplate, KindFlow:
		if !itemIDPattern.MatchString(id) {
			return Ref{}, fmt.Sprintf("rule %q has an invalid id %q in Applies to:", ruleID, id)
		}
	case KindTokens:
		if !namePattern.MatchString(id) {
			return Ref{}, fmt.Sprintf("rule %q has an invalid id %q in Applies to:", ruleID, id)
		}
	default:
		return Ref{}, fmt.Sprintf("rule %q has an unknown kind %q in Applies to:; use component, template, flow or tokens", ruleID, string(kind))
	}
	return Ref{Kind: kind, ID: id, Line: line}, ""
}
```

In `parseRules`, replace the entry loop (from `for _, e := range entries {` to the end of the function) with this version. The line-splitting loop above it is unchanged.

```go
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

		before := len(problems)
		var (
			statement  []string
			reason     string
			reasonLine int
			reasons    int
			appliesTo  int
			refs       []Ref
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
			if strings.HasPrefix(t, "Applies to:") {
				appliesTo++
				if appliesTo > 1 {
					add(l.no, "rule %q has more than one Applies to: line", e.id)
					continue
				}
				list := strings.TrimSpace(strings.TrimPrefix(t, "Applies to:"))
				if list == "" {
					add(l.no, "rule %q has an empty Applies to: line", e.id)
					continue
				}
				for _, entry := range strings.Split(list, ",") {
					ref, msg := parseRef(e.id, entry, l.no)
					if msg != "" {
						add(l.no, "%s", msg)
						continue
					}
					refs = append(refs, ref)
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
			AppliesTo: refs,
			Valid:     len(problems) == before,
		})
	}
	return rules, problems
}
```

Note the `rules.go` imports must include `fmt`, `regexp`, `strings` (already present).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/designsystem -v`
Expected: PASS for all tests, including every `TestParseRulesAppliesTo` case and the unchanged `TestParseRules`.

---

### Task 3: Lookup methods on DesignSystem

**Files:**
- Create: `internal/designsystem/lookup.go`
- Create: `internal/designsystem/lookup_test.go`

**Interfaces:**
- Consumes: `DesignSystem`, `Token`, `Item`, `Rule`, `Kind`, `Ref` from Tasks 1–2 and the existing `load.go`.
- Produces (all pure, no I/O):
  - `func (ds DesignSystem) Category(name string) ([]Token, bool)` — tokens of the category in file order; `bool` is true when at least one token exists.
  - `func (ds DesignSystem) Categories() []string` — distinct category names, sorted.
  - `func (ds DesignSystem) Item(kind Kind, id string) (Item, bool)` — kinds `KindComponent`, `KindTemplate`, `KindFlow`; any other kind returns false.
  - `func (ds DesignSystem) ItemIDs(kind Kind) []string` — ids of that kind, sorted; empty (not nil-error) when none.
  - `func (ds DesignSystem) Linked(kind Kind, id string) []Rule` — valid rules with at least one ref equal to `kind`/`id`, each rule once, in `Rules` order.

**Model:** haiku

- [ ] **Step 1: Write the failing test**

Create `internal/designsystem/lookup_test.go`:

```go
package designsystem

import (
	"reflect"
	"strings"
	"testing"
)

// sampleSystem is a hand-built pack. Rule a is valid and linked twice to button; b is invalid; c is valid and linked to color and button.
func sampleSystem() DesignSystem {
	return DesignSystem{
		Tokens: []Token{
			{Category: "space", Name: "unit", Value: "4"},
			{Category: "color", Name: "ink", Value: "#1d2433"},
			{Category: "color", Name: "accent", Value: "#2f6fed"},
		},
		Rules: []Rule{
			{ID: "a", Valid: true, AppliesTo: []Ref{{Kind: KindComponent, ID: "button"}, {Kind: KindComponent, ID: "button"}}},
			{ID: "b", Valid: false, AppliesTo: []Ref{{Kind: KindComponent, ID: "button"}}},
			{ID: "c", Valid: true, AppliesTo: []Ref{{Kind: KindTokens, ID: "color"}, {Kind: KindComponent, ID: "button"}}},
		},
		Components: []Item{{ID: "card", Title: "Card"}, {ID: "button", Title: "Button"}},
		Flows:      []Item{{ID: "first-run", Title: "First run"}},
	}
}

func ruleIDs(rs []Rule) []string {
	out := []string{}
	for _, r := range rs {
		out = append(out, r.ID)
	}
	return out
}

func TestLookups(t *testing.T) {
	ds := sampleSystem()

	t.Run("Category keeps file order and reports missing", func(t *testing.T) {
		got, ok := ds.Category("color")
		if !ok || len(got) != 2 || got[0].Name != "ink" || got[1].Name != "accent" {
			t.Errorf("Category(color) = %v, %t", got, ok)
		}
		if got, ok := ds.Category("motion"); ok || len(got) != 0 {
			t.Errorf("Category(motion) = %v, %t, want none", got, ok)
		}
	})

	t.Run("Categories sorted and unique", func(t *testing.T) {
		if got, want := ds.Categories(), []string{"color", "space"}; !reflect.DeepEqual(got, want) {
			t.Errorf("Categories() = %v, want %v", got, want)
		}
	})

	t.Run("Item finds by kind and id", func(t *testing.T) {
		if it, ok := ds.Item(KindComponent, "button"); !ok || it.Title != "Button" {
			t.Errorf("Item(component, button) = %v, %t", it, ok)
		}
		if _, ok := ds.Item(KindFlow, "button"); ok {
			t.Error("Item(flow, button) found a component")
		}
		if _, ok := ds.Item(KindTokens, "color"); ok {
			t.Error("Item(tokens, color) found something")
		}
	})

	t.Run("ItemIDs sorted per kind", func(t *testing.T) {
		if got, want := ds.ItemIDs(KindComponent), []string{"button", "card"}; !reflect.DeepEqual(got, want) {
			t.Errorf("ItemIDs(component) = %v, want %v", got, want)
		}
		if got := ds.ItemIDs(KindTokens); len(got) != 0 {
			t.Errorf("ItemIDs(tokens) = %v, want none", got)
		}
	})

	t.Run("Linked returns valid rules once, in rules order", func(t *testing.T) {
		if got, want := ruleIDs(ds.Linked(KindComponent, "button")), []string{"a", "c"}; !reflect.DeepEqual(got, want) {
			t.Errorf("Linked(component, button) = %v, want %v", got, want)
		}
		if got, want := ruleIDs(ds.Linked(KindTokens, "color")), []string{"c"}; !reflect.DeepEqual(got, want) {
			t.Errorf("Linked(tokens, color) = %v, want %v", got, want)
		}
		if got := ds.Linked(KindComponent, "card"); len(got) != 0 {
			t.Errorf("Linked(component, card) = %v, want none", ruleIDs(got))
		}
	})
}

func TestItemIDsDoesNotShareState(t *testing.T) {
	ds := sampleSystem()
	ids := ds.ItemIDs(KindComponent)
	ids[0] = "changed"
	if !strings.HasPrefix(ds.Components[0].ID, "card") {
		t.Errorf("ItemIDs mutated the pack: %v", ds.Components)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/designsystem -run 'TestLookups|TestItemIDsDoesNotShareState' -v`
Expected: FAIL to compile with `ds.Category undefined` (and the other lookup methods).

- [ ] **Step 3: Write minimal implementation**

Create `internal/designsystem/lookup.go`:

```go
package designsystem

import "slices"

// Category returns the tokens of the named category in file order, and whether there is at least one.
func (ds DesignSystem) Category(name string) ([]Token, bool) {
	var out []Token
	for _, t := range ds.Tokens {
		if t.Category == name {
			out = append(out, t)
		}
	}
	return out, len(out) > 0
}

// Categories returns the distinct token category names, sorted.
func (ds DesignSystem) Categories() []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range ds.Tokens {
		if !seen[t.Category] {
			seen[t.Category] = true
			out = append(out, t.Category)
		}
	}
	slices.Sort(out)
	return out
}

// itemsOf returns the items of a kind, or nil for KindTokens and unknown kinds.
func (ds DesignSystem) itemsOf(kind Kind) []Item {
	switch kind {
	case KindComponent:
		return ds.Components
	case KindTemplate:
		return ds.Templates
	case KindFlow:
		return ds.Flows
	}
	return nil
}

// Item returns the item of the given kind with the given id.
func (ds DesignSystem) Item(kind Kind, id string) (Item, bool) {
	for _, it := range ds.itemsOf(kind) {
		if it.ID == id {
			return it, true
		}
	}
	return Item{}, false
}

// ItemIDs returns the ids of the items of a kind, sorted.
func (ds DesignSystem) ItemIDs(kind Kind) []string {
	items := ds.itemsOf(kind)
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.ID)
	}
	slices.Sort(out)
	return out
}

// Linked returns the valid rules with at least one ref to kind:id, each rule once, in rules.md order.
func (ds DesignSystem) Linked(kind Kind, id string) []Rule {
	var out []Rule
	for _, r := range ds.Rules {
		if !r.Valid {
			continue
		}
		for _, ref := range r.AppliesTo {
			if ref.Kind == kind && ref.ID == id {
				out = append(out, r)
				break
			}
		}
	}
	return out
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/designsystem -v`
Expected: PASS for all tests, including `TestLookups` and `TestItemIDsDoesNotShareState`.

---

### Task 4: Cross-reference check in Load

**Files:**
- Modify: `internal/designsystem/load.go` (call `checkRefs` at the end of `Load`; add `checkRefs`, `refExists`, `refLabel`)
- Test: `internal/designsystem/load_test.go`

**Interfaces:**
- Consumes: `DesignSystem.Category`, `DesignSystem.Item` from Task 3; `Ref`, `Rule.AppliesTo` from Task 1.
- Produces: `func checkRefs(pack DesignSystem) []Problem`; `Load` appends its problems after the flows problems.

**Model:** haiku

- [ ] **Step 1: Write the failing test**

Append to `internal/designsystem/load_test.go`:

```go
func TestLoadReferences(t *testing.T) {
	base := map[string]string{
		"design-system/point-of-view.md":      "# Point of view\n\nText.\n",
		"design-system/tokens.json":           `{"color": {"ink": "#000"}}`,
		"design-system/components/button.md":  "# Button\n",
		"design-system/templates/page.md":     "# Page\n",
		"design-system/flows/first-run.md":    "# First run\n",
	}
	tests := []struct {
		name         string
		rules        string
		wantProblems []wantProblem
	}{
		{
			name:  "all targets exist",
			rules: "## a.b\nStatement.\nReason: r.\nApplies to: tokens:color, component:button, template:page, flow:first-run\n",
		},
		{
			name:  "unknown component",
			rules: "## a.b\nStatement.\nReason: r.\nApplies to: component:card\n",
			wantProblems: []wantProblem{
				{4, `rule "a.b" applies to unknown component "card"`},
			},
		},
		{
			name:  "unknown token category",
			rules: "# Rules\n\n## a.b\nStatement.\nReason: r.\nApplies to: tokens:space\n",
			wantProblems: []wantProblem{
				{5, `rule "a.b" applies to unknown token category "space"`},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			files := map[string]string{"design-system/rules.md": tt.rules}
			for k, v := range base {
				files[k] = v
			}
			writeTree(t, root, files)
			pack, problems := Load(root)
			checkProblems(t, problems, tt.wantProblems)
			if len(tt.wantProblems) == 0 && len(pack.Rules) != 1 {
				t.Fatalf("got %d rules, want 1", len(pack.Rules))
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/designsystem -run TestLoadReferences -v`
Expected: FAIL. The unknown-target cases report no problems, so `checkProblems` fails on the count.

- [ ] **Step 3: Write minimal implementation**

In `internal/designsystem/load.go`, in `Load`, replace the final `return pack, problems` with:

```go
	problems = append(problems, checkRefs(pack)...)
	return pack, problems
```

Add to `load.go` (add `"fmt"` to its imports):

```go
// checkRefs reports every rule ref that names an item or token category the pack does not have.
func checkRefs(pack DesignSystem) []Problem {
	var problems []Problem
	for _, r := range pack.Rules {
		for _, ref := range r.AppliesTo {
			if refExists(pack, ref) {
				continue
			}
			problems = append(problems, Problem{
				File:    RulesPath,
				Line:    ref.Line,
				Message: fmt.Sprintf("rule %q applies to unknown %s %q", r.ID, refLabel(ref.Kind), ref.ID),
			})
		}
	}
	return problems
}

func refExists(pack DesignSystem, ref Ref) bool {
	if ref.Kind == KindTokens {
		_, ok := pack.Category(ref.ID)
		return ok
	}
	_, ok := pack.Item(ref.Kind, ref.ID)
	return ok
}

func refLabel(kind Kind) string {
	if kind == KindTokens {
		return "token category"
	}
	return string(kind)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/designsystem -v`
Expected: PASS for all tests, including `TestLoadReferences` and the existing `TestLoadFixtureIsValid` (the fixture has no `Applies to:` lines yet, so it still passes).

---

### Task 5: Link the fixture's rules

**Files:**
- Modify: `testdata/project/design-system/rules.md`
- Test: `internal/designsystem/lookup_test.go` (append)

**Interfaces:**
- Consumes: `Load`, `Linked`, `Category` from Tasks 3–4.
- Produces: the fixture with two linked rules (spec D17).

**Model:** haiku

- [ ] **Step 1: Write the failing test**

Append to `internal/designsystem/lookup_test.go`:

```go
func TestFixtureLinks(t *testing.T) {
	pack, problems := Load("../../testdata/project")
	if len(problems) != 0 {
		t.Fatalf("fixture has problems: %v", problems)
	}
	if got, want := ruleIDs(pack.Linked(KindComponent, "button")), []string{"color.tokens-only"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Linked(component, button) = %v, want %v", got, want)
	}
	if got, want := ruleIDs(pack.Linked(KindTokens, "space")), []string{"spacing.unit-multiples"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Linked(tokens, space) = %v, want %v", got, want)
	}
	if got := pack.Linked(KindFlow, "first-run"); len(got) != 0 {
		t.Errorf("Linked(flow, first-run) = %v, want none", ruleIDs(got))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/designsystem -run TestFixtureLinks -v`
Expected: FAIL: `Linked(component, button)` is empty because the fixture has no `Applies to:` lines.

- [ ] **Step 3: Write minimal implementation**

Replace the full contents of `testdata/project/design-system/rules.md` with:

```markdown
# Rules

## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
Applies to: tokens:color, component:button

## spacing.unit-multiples
Spacing between elements is a multiple of the space unit token.
Reason: Off-grid spacing is the quickest way for a screen to look unplanned.
Applies to: tokens:space, template:settings-page
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/designsystem -v && go run ./cmd/tenet validate --dir testdata/project`
Expected: PASS for the package; validate prints `design system valid: 4 tokens, 2 rules, 1 components, 1 templates, 1 flows` and exits 0. (The summary's pluralisation is slice 2's existing behaviour; do not change it here.)

---

### Task 6: Pure response rendering

**Files:**
- Create: `internal/server/delivery.go`
- Create: `internal/server/delivery_test.go`

**Interfaces:**
- Consumes: `designsystem.Rule`, `designsystem.Problem`, `designsystem.Kind`, `designsystem.ComponentsDir`, `TemplatesDir`, `FlowsDir`.
- Produces:
  - `const rulesNoteText` = `Note: design-system/rules.md has problems, so some linked rules may be missing. Run tenet validate.`
  - `func deliveryText(lead string, rules []designsystem.Rule, rulesNote bool) string`
  - `type kindName struct { singular, plural, none, hint, dir string; kind designsystem.Kind }`
  - `var componentKind, templateKind, flowKind, tokenKind kindName`
  - `func notFoundText(k kindName, id string, available []string) string`
  - `func invalidText(k kindName, id string, p designsystem.Problem) string`

**Model:** haiku

- [ ] **Step 1: Write the failing test**

Create `internal/server/delivery_test.go`:

```go
package server

import (
	"testing"

	"github.com/tenet-ui/tenet/internal/designsystem"
)

func TestDeliveryText(t *testing.T) {
	rule := designsystem.Rule{ID: "a", Statement: "S.", Reason: "R.", Valid: true}
	rule2 := designsystem.Rule{ID: "b", Statement: "S2.", Reason: "R2."}
	tests := []struct {
		name  string
		lead  string
		rules []designsystem.Rule
		note  bool
		want  string
	}{
		{
			name: "lead only",
			lead: "x",
			want: "x\n",
		},
		{
			name: "trailing newlines in lead are trimmed",
			lead: "x\n\n",
			want: "x\n",
		},
		{
			name:  "one rule",
			lead:  "x",
			rules: []designsystem.Rule{rule},
			want:  "x\n\n## Linked rules\n\n### a\nS.\nReason: R.\n",
		},
		{
			name:  "two rules keep their order",
			lead:  "x",
			rules: []designsystem.Rule{rule, rule2},
			want:  "x\n\n## Linked rules\n\n### a\nS.\nReason: R.\n\n### b\nS2.\nReason: R2.\n",
		},
		{
			name: "rules note only",
			lead: "x",
			note: true,
			want: "x\n\n" + rulesNoteText + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deliveryText(tt.lead, tt.rules, tt.note); got != tt.want {
				t.Errorf("deliveryText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNotFoundText(t *testing.T) {
	tests := []struct {
		name      string
		k         kindName
		available []string
		want      string
	}{
		{
			name:      "lists available components",
			k:         componentKind,
			available: []string{"button", "card"},
			want:      `component "x" not found. Available components: button, card.`,
		},
		{
			name: "no components yet",
			k:    componentKind,
			want: `component "x" not found. The design system has no components yet; add Markdown files under design-system/components/.`,
		},
		{
			name:      "token categories",
			k:         tokenKind,
			available: []string{"color", "space"},
			want:      `token category "x" not found. Available token categories: color, space.`,
		},
		{
			name: "no tokens yet",
			k:    tokenKind,
			want: `token category "x" not found. The design system has no tokens yet; add categories to design-system/tokens.json.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := notFoundText(tt.k, "x", tt.available); got != tt.want {
				t.Errorf("notFoundText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInvalidText(t *testing.T) {
	p := designsystem.Problem{File: "design-system/components/bad.md", Line: 1, Message: `first line must be a "# <title>" heading`}
	want := `component "bad" is not valid: design-system/components/bad.md:1: first line must be a "# <title>" heading. Run tenet validate.`
	if got := invalidText(componentKind, "bad", p); got != want {
		t.Errorf("invalidText() = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/server -run 'TestDeliveryText|TestNotFoundText|TestInvalidText' -v`
Expected: FAIL to compile (`deliveryText`, `kindName`, etc. undefined).

- [ ] **Step 3: Write minimal implementation**

Create `internal/server/delivery.go`:

```go
package server

import (
	"fmt"
	"strings"

	"github.com/tenet-ui/tenet/internal/designsystem"
)

const rulesNoteText = "Note: design-system/rules.md has problems, so some linked rules may be missing. Run tenet validate."

// deliveryText builds the text of a delivery result: the lead text, then the linked
// rules in the rules.md format (only if there are any), then the rules note (only if asked).
// This is the one place that defines the shape of a delivery result.
func deliveryText(lead string, rules []designsystem.Rule, rulesNote bool) string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(lead, " \t\r\n"))
	b.WriteString("\n")
	if len(rules) > 0 {
		b.WriteString("\n## Linked rules\n")
		for _, r := range rules {
			fmt.Fprintf(&b, "\n### %s\n%s\nReason: %s\n", r.ID, r.Statement, r.Reason)
		}
	}
	if rulesNote {
		b.WriteString("\n" + rulesNoteText + "\n")
	}
	return b.String()
}

// kindName describes one kind of delivery target for messages and file lookups.
type kindName struct {
	singular string
	plural   string
	none     string // used in "has no <none> yet"
	hint     string // what to add to get one
	dir      string // item folder; empty for tokens
	kind     designsystem.Kind
}

var (
	componentKind = kindName{"component", "components", "components", "add Markdown files under design-system/components/", designsystem.ComponentsDir, designsystem.KindComponent}
	templateKind  = kindName{"template", "templates", "templates", "add Markdown files under design-system/templates/", designsystem.TemplatesDir, designsystem.KindTemplate}
	flowKind      = kindName{"flow", "flows", "flows", "add Markdown files under design-system/flows/", designsystem.FlowsDir, designsystem.KindFlow}
	tokenKind     = kindName{"token category", "token categories", "tokens", "add categories to design-system/tokens.json", "", designsystem.KindTokens}
)

// notFoundText names the missing id and lists what exists, or says the kind has none yet.
func notFoundText(k kindName, id string, available []string) string {
	if len(available) == 0 {
		return fmt.Sprintf("%s %q not found. The design system has no %s yet; %s.", k.singular, id, k.none, k.hint)
	}
	return fmt.Sprintf("%s %q not found. Available %s: %s.", k.singular, id, k.plural, strings.Join(available, ", "))
}

// invalidText reports that the requested item exists on disk but its file has a problem.
func invalidText(k kindName, id string, p designsystem.Problem) string {
	return fmt.Sprintf("%s %q is not valid: %s. Run tenet validate.", k.singular, id, p.String())
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/server -v`
Expected: PASS for the new tests and the existing slice 1 tests.

---

### Task 7: Four delivery tools, registered and tested

**Files:**
- Modify: `internal/server/delivery.go` (add input types, handlers, helpers)
- Modify: `internal/server/server.go` (register four tools in `New`)
- Modify: `internal/server/server_test.go` (replace the one-tool list test with the five-tool test)
- Modify: `internal/server/delivery_test.go` (append tool-level tests)

**Interfaces:**
- Consumes: everything from Tasks 1–6; `designsystem.Load`, `designsystem.DirName`, `designsystem.TokensPath`, `designsystem.RulesPath`.
- Produces:
  - `type GetTokensInput struct { Category string }` (json `category`, required)
  - `type GetItemInput struct { ID string }` (json `id`, required)
  - `func deliverItem(root string, k kindName, id string) *mcp.CallToolResult`
  - `func deliverTokens(root, category string) *mcp.CallToolResult`
  - `func problemFor(problems []designsystem.Problem, file string) (designsystem.Problem, bool)`
  - `func textResult(s string) *mcp.CallToolResult`, `func errorText(s string) *mcp.CallToolResult`

**Model:** sonnet

- [ ] **Step 1: Write the failing tests**

Append to `internal/server/delivery_test.go`. Add these imports to the file's import block: `"context"`, `"encoding/json"`, `"os"`, `"path/filepath"`, `"sort"`, `"strings"`, `"testing"`, `"github.com/modelcontextprotocol/go-sdk/mcp"`, and keep `designsystem`.

```go
const fixtureRoot = "../../testdata/project"

// writeDesignSystem writes files under root, creating folders as needed.
func writeDesignSystem(t *testing.T, root string, files map[string]string) {
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

// callTool calls a tool on the session and returns its single text result.
func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	return res
}

func TestGetComponentFixtureExact(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_component", map[string]any{"id": "button"})
	if res.IsError {
		t.Fatalf("IsError, text: %s", textOf(t, res))
	}
	want := "# Button\n\nThe primary action on a surface. Use one per view.\n\n## Linked rules\n\n### color.tokens-only\nComponents take colours from tokens, never from raw hex values.\nReason: A raw value drifts from the palette the moment one screen is edited alone.\n"
	if got := textOf(t, res); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestGetFlowFixtureHasNoLinkedRules(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_flow", map[string]any{"id": "first-run"})
	if res.IsError {
		t.Fatalf("IsError, text: %s", textOf(t, res))
	}
	if got := textOf(t, res); strings.Contains(got, "Linked rules") || !strings.HasPrefix(got, "# First run\n") {
		t.Errorf("got %q, want the flow body and no linked rules", got)
	}
}

func TestGetTokensFixtureExact(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_tokens", map[string]any{"category": "color"})
	if res.IsError {
		t.Fatalf("IsError, text: %s", textOf(t, res))
	}
	want := "color.ink: #1d2433\ncolor.paper: #fbfaf7\ncolor.accent: #2f6fed\n\n## Linked rules\n\n### color.tokens-only\nComponents take colours from tokens, never from raw hex values.\nReason: A raw value drifts from the palette the moment one screen is edited alone.\n"
	if got := textOf(t, res); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestGetTokensNumberValue(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_tokens", map[string]any{"category": "space"})
	if res.IsError {
		t.Fatalf("IsError, text: %s", textOf(t, res))
	}
	if got := textOf(t, res); !strings.HasPrefix(got, "space.unit: 4\n") {
		t.Errorf("got %q, want it to start with space.unit: 4", got)
	}
}

func TestNotFoundListsAvailable(t *testing.T) {
	session := connect(t, fixtureRoot)
	tests := []struct {
		tool string
		args map[string]any
		want string
	}{
		{"get_component", map[string]any{"id": "card"}, `component "card" not found. Available components: button.`},
		{"get_template", map[string]any{"id": "missing"}, `template "missing" not found. Available templates: settings-page.`},
		{"get_flow", map[string]any{"id": "x"}, `flow "x" not found. Available flows: first-run.`},
		{"get_tokens", map[string]any{"category": "motion"}, `token category "motion" not found. Available token categories: color, space.`},
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			res := callTool(t, session, tt.tool, tt.args)
			if !res.IsError {
				t.Fatal("want IsError")
			}
			if got := textOf(t, res); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmptyIDListsAvailable(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_component", map[string]any{"id": ""})
	if !res.IsError {
		t.Fatal("want IsError")
	}
	if got := textOf(t, res); !strings.HasPrefix(got, `component "" not found. Available components: button.`) {
		t.Errorf("got %q", got)
	}
}

func TestNoItemsOfKindSaysSo(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "# Rules\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "get_component", map[string]any{"id": "button"})
	want := `component "button" not found. The design system has no components yet; add Markdown files under design-system/components/.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestInvalidItemReportsProblem(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md":    "# Point of view\n\nText.\n",
		"design-system/tokens.json":         "{}",
		"design-system/rules.md":            "# Rules\n",
		"design-system/components/bad.md":   "Bad\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "get_component", map[string]any{"id": "bad"})
	want := `component "bad" is not valid: design-system/components/bad.md:1: first line must be a "# <title>" heading. Run tenet validate.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestMissingDesignSystemIsToolError(t *testing.T) {
	session := connect(t, t.TempDir())
	for _, tt := range []struct {
		tool string
		args map[string]any
	}{
		{"get_component", map[string]any{"id": "button"}},
		{"get_template", map[string]any{"id": "x"}},
		{"get_flow", map[string]any{"id": "x"}},
		{"get_tokens", map[string]any{"category": "color"}},
	} {
		res := callTool(t, session, tt.tool, tt.args)
		if !res.IsError || !strings.Contains(textOf(t, res), "run tenet init") {
			t.Errorf("%s: got IsError=%t %q, want run tenet init", tt.tool, res.IsError, textOf(t, res))
		}
	}
}

func TestEditsVisibleWithoutRestart(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md":   "# Point of view\n\nText.\n",
		"design-system/tokens.json":        "{}",
		"design-system/rules.md":           "# Rules\n",
		"design-system/components/button.md": "# Button\n\nOld.\n",
	})
	session := connect(t, root)
	if got := textOf(t, callTool(t, session, "get_component", map[string]any{"id": "button"})); !strings.Contains(got, "Old.") {
		t.Fatalf("first call got %q", got)
	}
	writeDesignSystem(t, root, map[string]string{"design-system/components/button.md": "# Button\n\nNew.\n"})
	if got := textOf(t, callTool(t, session, "get_component", map[string]any{"id": "button"})); !strings.Contains(got, "New.") {
		t.Errorf("second call got %q, want the edited body", got)
	}
}

func TestRulesProblemNote(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md":    "# Point of view\n\nText.\n",
		"design-system/tokens.json":         "{}",
		"design-system/rules.md":            "## a.b\nStatement.\nApplies to: component:button\n",
		"design-system/components/button.md": "# Button\n\nBody.\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "get_component", map[string]any{"id": "button"})
	want := "# Button\n\nBody.\n\n" + rulesNoteText + "\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestResultsUnder4KB(t *testing.T) {
	session := connect(t, fixtureRoot)
	for _, tt := range []struct {
		tool string
		args map[string]any
	}{
		{"get_component", map[string]any{"id": "button"}},
		{"get_template", map[string]any{"id": "settings-page"}},
		{"get_flow", map[string]any{"id": "first-run"}},
		{"get_tokens", map[string]any{"category": "color"}},
	} {
		res := callTool(t, session, tt.tool, tt.args)
		if n := len(textOf(t, res)); n >= 4096 {
			t.Errorf("%s returned %d bytes, want under 4096", tt.tool, n)
		}
	}
}

func TestDeliveryToolSchemas(t *testing.T) {
	session := connect(t, fixtureRoot)
	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	wantRequired := map[string]string{"get_component": "id", "get_template": "id", "get_flow": "id", "get_tokens": "category"}
	for _, tool := range res.Tools {
		field, ok := wantRequired[tool.Name]
		if !ok {
			continue
		}
		raw, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatal(err)
		}
		var schema struct {
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatal(err)
		}
		if !sort.SearchStringsAreSorted(schema.Required, field) && !containsString(schema.Required, field) {
			t.Errorf("%s schema required = %v, want %q (schema: %s)", tool.Name, schema.Required, field, raw)
		}
	}
}

func containsString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
```

Replace the `sort.SearchStringsAreSorted` line with the plain check so `sort` is not needed: `if !containsString(schema.Required, field) {`. Remove `"sort"` from the imports.

In `internal/server/server_test.go`, replace the whole `TestListsOnlyGetPointOfView` function with:

```go
func TestListsFiveTools(t *testing.T) {
	session := connect(t, fixtureRoot)

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	var names []string
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
		raw, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatalf("marshal input schema: %v", err)
		}
		var schema map[string]any
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("unmarshal input schema %s: %v", raw, err)
		}
		if schema["type"] != "object" {
			t.Errorf("%s input schema type = %v, want \"object\"", tool.Name, schema["type"])
		}
	}
	want := []string{"get_component", "get_flow", "get_point_of_view", "get_template", "get_tokens"}
	if !reflect.DeepEqual(sortedCopy(names), want) {
		t.Fatalf("tools = %v, want %v", names, want)
	}
}

func sortedCopy(ss []string) []string {
	out := append([]string(nil), ss...)
	slices.Sort(out)
	return out
}
```

Add `"reflect"` and `"slices"` to `server_test.go` imports.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/server -v`
Expected: FAIL. The new tools are not registered, so `TestListsFiveTools` reports the wrong tool list and the tool-level tests get "unknown tool" errors.

- [ ] **Step 3: Write minimal implementation**

Append to `internal/server/delivery.go`, and add `"context"` is not needed here; add the mcp import so the file reads `"fmt"`, `"strings"`, `"github.com/modelcontextprotocol/go-sdk/mcp"`, `"github.com/tenet-ui/tenet/internal/designsystem"`:

```go
// GetTokensInput is the input of get_tokens.
type GetTokensInput struct {
	Category string `json:"category" jsonschema:"the token category, for example color"`
}

// GetItemInput is the input of get_component, get_template and get_flow.
type GetItemInput struct {
	ID string `json:"id" jsonschema:"the item id, the file name without .md"`
}

// deliverItem loads the pack and returns the item with its linked rules.
func deliverItem(root string, k kindName, id string) *mcp.CallToolResult {
	pack, problems := designsystem.Load(root)
	if p, ok := problemFor(problems, designsystem.DirName); ok {
		return errorText(p.String())
	}
	if item, ok := pack.Item(k.kind, id); ok {
		_, note := problemFor(problems, designsystem.RulesPath)
		return textResult(deliveryText(item.Body, pack.Linked(k.kind, id), note))
	}
	file := k.dir + "/" + id + ".md"
	return errorText(missingText(problems, k, file, id, pack.ItemIDs(k.kind)))
}

// deliverTokens loads the pack and returns one line per token of the category, with its linked rules.
func deliverTokens(root, category string) *mcp.CallToolResult {
	pack, problems := designsystem.Load(root)
	if p, ok := problemFor(problems, designsystem.DirName); ok {
		return errorText(p.String())
	}
	if tokens, ok := pack.Category(category); ok {
		lines := make([]string, 0, len(tokens))
		for _, t := range tokens {
			lines = append(lines, fmt.Sprintf("%s.%s: %s", t.Category, t.Name, t.Value))
		}
		_, note := problemFor(problems, designsystem.RulesPath)
		return textResult(deliveryText(strings.Join(lines, "\n"), pack.Linked(designsystem.KindTokens, category), note))
	}
	return errorText(missingText(problems, tokenKind, designsystem.TokensPath, category, pack.Categories()))
}

// missingText prefers the file's own problem over a plain not-found message.
func missingText(problems []designsystem.Problem, k kindName, file, id string, available []string) string {
	if p, ok := problemFor(problems, file); ok {
		return invalidText(k, id, p)
	}
	return notFoundText(k, id, available)
}

// problemFor returns the first problem reported for file.
func problemFor(problems []designsystem.Problem, file string) (designsystem.Problem, bool) {
	for _, p := range problems {
		if p.File == file {
			return p, true
		}
	}
	return designsystem.Problem{}, false
}

func textResult(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

func errorText(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}
```

In `internal/server/server.go`, inside `New`, after the existing `mcp.AddTool(... "get_point_of_view" ...)` call and before `return s`, add:

```go
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_tokens",
		Description: "Returns the design tokens of one category (for example color) and the valid rules linked to it. Pass the category name, which is the first part of a token's name.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in GetTokensInput) (*mcp.CallToolResult, any, error) {
		return deliverTokens(root, in.Category), nil, nil
	})
	for _, d := range []struct {
		name, desc string
		k          kindName
	}{
		{"get_component", "Returns the design component with the given id (its file name without .md) and the valid rules linked to it.", componentKind},
		{"get_template", "Returns the design template with the given id (its file name without .md) and the valid rules linked to it.", templateKind},
		{"get_flow", "Returns the design flow with the given id (its file name without .md) and the valid rules linked to it.", flowKind},
	} {
		mcp.AddTool(s, &mcp.Tool{Name: d.name, Description: d.desc}, func(_ context.Context, _ *mcp.CallToolRequest, in GetItemInput) (*mcp.CallToolResult, any, error) {
			return deliverItem(root, d.k, in.ID), nil, nil
		})
	}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/server -v && go test ./...`
Expected: PASS for every test in every package. If `TestDeliveryToolSchemas` fails because the SDK does not emit a `required` list for these fields, stop and report BLOCKED with the schema output; do not remove the `json` tag or the check.

---

### Task 8: README

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: the final tool names, arguments and messages from Tasks 6–7.
- Produces: user documentation for `Applies to:` and the four tools.

**Model:** sonnet

- [ ] **Step 1: Write the failing check**

Run: `rg -n "Applies to|get_tokens|get_component" README.md`
Expected: no output (the README does not mention any of them yet). Treat that as the failing check.

- [ ] **Step 2: Run check to verify it fails**

Covered by Step 1: no matches.

- [ ] **Step 3: Write minimal implementation**

Edit `README.md`:

1. In the intro paragraph, change "It has one MCP tool, `get_point_of_view`. The design-system tools arrive in a later slice." to: "It has five MCP tools: `get_point_of_view`, and four delivery tools that return one item at a time (see [Delivery tools](#delivery-tools))."
2. In the rules section (the paragraph starting "**`rules.md`** holds one rule per `##` heading"), after the example and the "A rules file with no rules is valid." sentence, add:

````markdown
A rule can name the items and token categories it applies to with one `Applies to:` line. Each entry is `kind:id`, where kind is `component`, `template`, `flow` or `tokens`, and for `tokens` the id is a category name:

```markdown
## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
Applies to: tokens:color, component:button
```

A rule without an `Applies to:` line is valid, but no delivery tool returns it. Every reference must name an item or category that exists; `validate` reports the ones that do not.
````

3. Add a new section before "## Check the stdio protocol":

````markdown
## Delivery tools

| Tool | Argument | Returns |
|---|---|---|
| `get_tokens` | `category`, for example `color` | one `category.name: value` line per token, then the rules linked to the category |
| `get_component` | `id`, the file name without `.md` | the file's text, then the rules linked to the component |
| `get_template` | `id` | the same, for `templates/` |
| `get_flow` | `id` | the same, for `flows/` |

A rule appears under `## Linked rules` only when it is valid and names the item or category. An unknown id returns an error that lists the ids that exist. An item whose file has a problem returns that problem. Each call reads the design system again, so edits show without restarting the server. Keep each result small: a single item and its rules should stay well under 4 KB.
````

4. In "## Check the stdio protocol", change the request in the JSON to call `get_component` with `{"id":"button"}` instead of `get_point_of_view`, and change the sentence "Stdout carries only JSON-RPC responses. The startup line, which includes the design-system summary and any problems, goes to stderr." to stay as is. Also change the example's `"arguments":{}` in that request to `"arguments":{"id":"button"}`.

- [ ] **Step 4: Run check to verify it passes**

Run: `rg -n "Applies to|get_tokens|get_component" README.md && ( printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"check","version":"0"}}}' '{"jsonrpc":"2.0","method":"notifications/initialized"}' '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_component","arguments":{"id":"button"}}}'; sleep 3 ) | go run ./cmd/tenet --dir testdata/project`
Expected: the `rg` lines match; the piped run prints a JSON-RPC response to stdout whose text starts with `# Button` and includes `## Linked rules`; the stderr startup line reports the design system summary.

---

## Final gates (after all tasks)

Run these from the repo root and read the output before claiming done:

1. `go test ./...`
2. `test -z "$(gofmt -l .)"`
3. `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`
4. `go run ./cmd/tenet validate --dir testdata/project` exits 0
5. Each acceptance criterion observed in the running app (see the `run` skill): four delivery calls on the fixture, one missing-id error, and one stdio check.

## Self-review

- **Spec coverage:** D1–D3 → Tasks 1–2, 4. D4–D5 → Tasks 1–2, 3 (`Linked` skips invalid). D6–D7 → Task 1. D8–D9 → Tasks 6–7. D10–D12 → Tasks 6–7 (`notFoundText`, `missingText`, `deliverItem`/`deliverTokens` DirName check). D13 → Task 7 (`Load` per call; `New(root)` unchanged). D14 → Tasks 6–7 (note). D15 → Tasks 3, 6–7 (designsystem stays MCP-free). D16 → Task 7 (`get_point_of_view` untouched; five-tool test). D17 → Task 5. D18–D19 → no change needed; Task 4 reuses existing `validate`. Acceptance 1 (under 4 KB) → `TestResultsUnder4KB`. Acceptance 2 (errors name and list) → `TestNotFoundListsAvailable`.
- **Placeholder scan:** none; every code step has its code.
- **Type consistency:** `Ref`, `Kind`, `Rule.AppliesTo`, `Rule.Valid`, `Item.Body` are defined in Task 1 and used unchanged after; `kindName` fields and the four `kindName` vars are defined in Task 6 and used in Task 7; `rulesNoteText` is defined in Task 6 and used in Task 7's test.
- **Ordering:** Task 3's `Category` is used by Task 4's `refExists`; Task 5 depends on Tasks 3–4; Tasks 6–7 depend on Tasks 1–5.
