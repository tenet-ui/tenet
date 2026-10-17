# Slice 5: Reference pack: unofficial Stripe-inspired system Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `packs/stripe-reference/`, a complete unofficial, Stripe-inspired design-system pack, plus a fixture that `critique` checks with no findings.

**Architecture:** The pack is data only: a `design-system/` folder read by the existing `designsystem.Load`, checked by the existing `designsystem.Critique`. No Go code under `internal/` changes. Tests in `internal/designsystem/pack_test.go` and one case in `internal/server/critique_test.go` pin the pack, the fixture and the patterns.

**Tech Stack:** Go (toolchain from `go.mod`), `github.com/modelcontextprotocol/go-sdk` (existing), Markdown and JSON content.

**Spec:** `docs/superpowers/specs/2026-10-18-slice-5-stripe-reference-pack-design.md`

## Global Constraints

- The pack ships no Stripe logos, wordmarks, font files or other brand assets (ADR-0004). Only `.md`, `.json` and `.html` files live under `packs/stripe-reference/`.
- Token values are the pack's own choices, not copied from research observations (spec R3).
- The pack's name is "Stripe-inspired" and it is labelled unofficial; it never claims Stripe endorsement (spec R9).
- No commit steps. The slice is committed once at the end (CLAUDE.md, "Commits"). Do not add a `Co-Authored-By` trailer.
- No change to `internal/designsystem/*.go` (non-test) or `internal/server/*.go` (non-test). Tool count stays at six.
- Run `rg` (ripgrep) for searches, not `grep`/`find` (your global CLAUDE.md, "Searching").

## Review Focus

1. A raw pixel value that the pack's rule allows, such as `margin: 0`, must not be flagged. Expected: no finding. Pinned by Task 1, `TestPackPatternsOnSnippets`.
2. A motion curve written as `cubic-bezier(...)` in a transition must not be flagged. Expected: no finding. Pinned by Task 1.
3. Uppercase hex colour such as `#FFFFFF` must be flagged. Expected: one `color.tokens-only` finding. Pinned by Task 1.
4. A raw colour fallback inside `var(--x, rgb(...))` must be flagged. Expected: one `color.tokens-only` finding. Pinned by Task 1.
5. A pixel value on a line after its property (`margin:` then `13px`) is not matched, because patterns are per line (spec C6, README limit). Expected: no finding. Pinned by Task 1.

---

### Task 1: Pack and pattern tests (failing first)

**Files:**
- Create: `internal/designsystem/pack_test.go`

**Interfaces:**
- Consumes: `Load(root string) (DesignSystem, []Problem)`, `Critique(source string, rules []Rule) []Finding`, `Checkable(rules []Rule) int`, existing in `internal/designsystem` (`DesignSystem`, `Finding.RuleID`, `Finding.Match`).
- Produces: test helpers `packRoot` (const) and `loadPack(t) DesignSystem`, used only in this file.

**Model: sonnet**

- [ ] **Step 1: Check the helper names are free**

Run: `rg -n "packRoot|loadPack" internal/designsystem`
Expected: no output. If a match exists, rename the new helpers to `stripePackRoot` and `loadStripePack` in every place this task writes them.

- [ ] **Step 2: Write the failing tests**

Create `internal/designsystem/pack_test.go` with exactly this content:

```go
package designsystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	packRoot    = "../../packs/stripe-reference"
	packFixture = "../../testdata/critique/pack-clean.html"
)

// loadPack loads the reference pack and fails the test when it has any problem.
func loadPack(t *testing.T) DesignSystem {
	t.Helper()
	pack, problems := Load(packRoot)
	if len(problems) != 0 {
		t.Fatalf("pack has problems: %v", problems)
	}
	return pack
}

func TestPackLoadsClean(t *testing.T) {
	pack := loadPack(t)
	if len(pack.Templates) != 1 || len(pack.Flows) != 1 || len(pack.Components) != 3 {
		t.Errorf("got %d templates, %d flows, %d components; want 1, 1, 3",
			len(pack.Templates), len(pack.Flows), len(pack.Components))
	}
	if len(pack.Rules) != 6 {
		t.Errorf("got %d rules, want 6", len(pack.Rules))
	}
	if got := Checkable(pack.Rules); got != 4 {
		t.Errorf("Checkable = %d, want 4", got)
	}
}

func TestPackShipsNoBrandAssets(t *testing.T) {
	err := filepath.WalkDir(packRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".md", ".json", ".html":
		default:
			t.Errorf("pack ships a non-text file: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPackReadmeStatesUnofficial(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(packRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(data))
	for _, want := range []string{"unofficial", "not endorsed by stripe", "docs/research/stripe-design-system.md"} {
		if !strings.Contains(text, want) {
			t.Errorf("README.md does not contain %q", want)
		}
	}
}

func TestPackFixtureHasNoFindings(t *testing.T) {
	pack := loadPack(t)
	data, err := os.ReadFile(packFixture)
	if err != nil {
		t.Fatal(err)
	}
	if got := Critique(string(data), pack.Rules); len(got) != 0 {
		t.Errorf("fixture has findings: %+v", got)
	}
}

func TestPackSeededViolationIsFound(t *testing.T) {
	pack := loadPack(t)
	data, err := os.ReadFile(packFixture)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data) + "<p style=\"color: #ff0000\">Seeded</p>\n"
	got := Critique(source, pack.Rules)
	if len(got) != 1 || got[0].RuleID != "color.tokens-only" || got[0].Match != "#ff0000" {
		t.Errorf("got %+v, want one color.tokens-only finding for #ff0000", got)
	}
}

func TestPackPatternsOnSnippets(t *testing.T) {
	pack := loadPack(t)
	tests := []struct {
		name   string
		source string
		want   []string // rule ID of each expected finding, in order
	}{
		{
			name:   "unitless zero margin is not a pixel value",
			source: "margin: 0;\n",
		},
		{
			name:   "a cubic-bezier curve in a transition is allowed",
			source: "transition: color 120ms cubic-bezier(0.25, 1, 0.5, 1);\n",
		},
		{
			name:   "uppercase hex is still a raw colour",
			source: "color: #FFFFFF;\n",
			want:   []string{"color.tokens-only"},
		},
		{
			name:   "a raw colour fallback inside var() is a raw colour",
			source: "color: var(--color-ink, rgb(0, 0, 0));\n",
			want:   []string{"color.tokens-only"},
		},
		{
			name:   "a pixel value on the next line is not matched (per-line limit)",
			source: "margin:\n13px\n",
		},
		{
			name:   "ease-in-out is flagged as a built-in keyword (documented)",
			source: "transition: color 1s ease-in-out;\n",
			want:   []string{"motion.no-builtin-easing"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ids []string
			for _, f := range Critique(tt.source, pack.Rules) {
				ids = append(ids, f.RuleID)
			}
			if !reflect.DeepEqual(ids, tt.want) {
				t.Errorf("rule IDs = %v, want %v", ids, tt.want)
			}
		})
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./internal/designsystem -run 'TestPack' -v`
Expected: FAIL. `TestPackLoadsClean` and the other `loadPack` tests fail with `pack has problems` mentioning `design-system not found` or `directory not found`. `TestPackShipsNoBrandAssets` fails with a `no such file or directory` error from the walk. The test file must compile; a compile error is not the expected failure.

- [ ] **Step 4: Stop here**

Tasks 2 to 5 make these tests pass. Do not edit this file again in later tasks.

---

### Task 2: Fixture page

**Files:**
- Create: `testdata/critique/pack-clean.html`

**Interfaces:**
- Consumes: none.
- Produces: the fixture file read by `packFixture` (Task 1) and by the server test (Task 6). Its checked token names are `color-ink`, `color-ink-soft`, `space-md`, `space-sm`, `space-xs`, `radius-md`, `motion-duration-fast`, `motion-standard`, which Task 3 defines.

**Model: haiku**

- [ ] **Step 1: Write the fixture with exactly this content**

```html
<!doctype html>
<head><style>
.card { color: var(--color-ink); margin: var(--space-md); border-radius: var(--radius-md); transition: color var(--motion-duration-fast) var(--motion-standard); }
.note { color: var(--color-ink-soft); padding: var(--space-sm) var(--space-md); gap: var(--space-xs); }
</style></head>
<body><p class="card">Hello</p></body>
```

- [ ] **Step 2: Check the fixture by reading it**

Run: `rg -n "#|rgba?\(|hsla?\(|px" testdata/critique/pack-clean.html`
Expected: no output. Any match means a raw value slipped in; fix the file.

- [ ] **Step 3: Run the fixture-only tests**

Run: `go test ./internal/designsystem -run 'TestPackFixtureHasNoFindings' -v`
Expected: still FAIL, with a `pack has problems` message, because the pack does not exist until Task 3. This is the expected state.

---

### Task 3: Tokens and rules

**Files:**
- Create: `packs/stripe-reference/design-system/tokens.json`
- Create: `packs/stripe-reference/design-system/rules.md`

**Interfaces:**
- Consumes: the token names used by the fixture (Task 2).
- Produces: categories `color`, `space`, `radius`, `motion`, `type`. Rule IDs `color.tokens-only`, `spacing.tokens-only`, `radius.tokens-only`, `motion.no-builtin-easing`, `button.one-primary`, `empty-state.title-states-missing`. Applies-to refs use the components and template from Task 4.

**Model: haiku**

- [ ] **Step 1: Write `tokens.json` with exactly this content**

```json
{
  "color": {
    "ink": "#0d1b2e",
    "ink-soft": "#3d4a5c",
    "ink-quiet": "#6b7a90",
    "paper": "#ffffff",
    "paper-subdued": "#f6f8fb",
    "border": "#dfe6ef",
    "brand": "#4f3bd6",
    "brand-hover": "#3f2fb3",
    "brand-soft": "#ece9ff",
    "focus": "#7b6ff0",
    "danger": "#c42b1c"
  },
  "space": {
    "unit": 8,
    "xs": 4,
    "sm": 8,
    "md": 16,
    "lg": 24,
    "xl": 32
  },
  "radius": {
    "xs": 2,
    "sm": 4,
    "md": 6,
    "lg": 16,
    "round": 9999
  },
  "motion": {
    "standard": "cubic-bezier(0.25, 1, 0.5, 1)",
    "duration-fast": "120ms",
    "duration-base": "300ms"
  },
  "type": {
    "family": "\"Inter\", system-ui, sans-serif",
    "weight-light": 300,
    "weight-bold": 500
  }
}
```

- [ ] **Step 2: Write `rules.md` with exactly this content**

```markdown
# Rules

## color.tokens-only
Components take colours from tokens, never from raw hex values or rgb() and hsl() functions.
Reason: A raw value drifts from the palette the moment one screen is edited alone (research section 1).
Applies to: tokens:color, component:button, component:text-field, template:payments-list
Forbid: #[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?)\(
Fix: Replace the raw colour with a colour token from design-system/tokens.json.

## spacing.tokens-only
Spacing between elements uses space tokens, not raw pixel values.
Reason: Off-grid spacing is the quickest way for a screen to look unplanned (research section 3, an 8px-based scale).
Applies to: tokens:space, component:button, component:text-field, template:payments-list
Forbid: \b(margin|padding|gap)\b[^;"]*\b\d+px
Fix: Replace the raw pixel value with a space token from design-system/tokens.json.

## radius.tokens-only
Corner radius comes from a radius token, not a raw pixel value.
Reason: The corner scale is small and fixed, so one raw value stands out (research section 3, radii of 2, 4, 6, 16 and 32px, and round).
Applies to: tokens:radius, component:button, component:text-field
Forbid: border-radius[^;"]*\b\d+px
Fix: Replace the raw pixel radius with a radius token from design-system/tokens.json.

## motion.no-builtin-easing
Transitions and animations use a cubic-bezier curve from the motion tokens, not a built-in easing keyword.
Reason: Built-in ease-in, ease-out and linear curves look generic next to a tuned curve (research section 4).
Applies to: tokens:motion, component:button
Forbid: (transition|animation)[^;"]*\b(ease-in|ease-out|linear)\b
Fix: Replace the easing keyword with the motion.standard token from design-system/tokens.json.

## button.one-primary
A view has at most one primary button; other actions use the secondary style.
Reason: Two primary actions compete, so the user cannot tell which one the screen expects (research section 7, checkout and onboarding patterns).
Applies to: component:button, template:payments-list

## empty-state.title-states-missing
An empty state's title states what is missing, as a short phrase ending in a period, and its action repeats the title's noun.
Reason: A title that promotes instead of stating the gap reads as marketing (research section 6, the Stripe Apps empty-state rules).
Applies to: component:empty-state, template:payments-list
```

- [ ] **Step 3: Check the JSON parses**

Run: `python3 -I -c "import json,sys; json.load(open('packs/stripe-reference/design-system/tokens.json')); print('ok')"`
Expected: `ok`

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/designsystem -run 'TestPack' -v`
Expected: still FAIL, because the components, template, flow and point-of-view file from Task 4 are missing, so `Applies to:` refs fail to resolve. This is the expected state; Task 4 fixes it.

---

### Task 4: Point of view, components, template and flow

**Files:**
- Create: `packs/stripe-reference/design-system/point-of-view.md`
- Create: `packs/stripe-reference/design-system/components/button.md`
- Create: `packs/stripe-reference/design-system/components/text-field.md`
- Create: `packs/stripe-reference/design-system/components/empty-state.md`
- Create: `packs/stripe-reference/design-system/templates/payments-list.md`
- Create: `packs/stripe-reference/design-system/flows/account-setup.md`

**Interfaces:**
- Consumes: the `Applies to:` refs in Task 3 (`component:button`, `component:text-field`, `component:empty-state`, `template:payments-list`).
- Produces: item files read by `Load`. Each carries a `Source:` line citing its research section (spec R9).

**Model: sonnet**

- [ ] **Step 1: Write `point-of-view.md` with exactly this content**

```markdown
# Point of view

A calm, precise system for payment-adjacent products. It favours clear hierarchy, light display type, quiet borders and one saturated accent, and it treats the customer's next step as the centre of every screen.

This is an unofficial, Stripe-inspired reference pack. It is not a Stripe product and is not endorsed by Stripe. Its observed sources are listed in `packs/stripe-reference/README.md`.
```

- [ ] **Step 2: Write `components/button.md` with exactly this content**

```markdown
# Button

The primary action on a surface. Use one per view; other actions use the secondary style.

Source: docs/research/stripe-design-system.md section 1 (colour tokens), section 2 (button type), section 3 (button height and radius).
```

- [ ] **Step 3: Write `components/text-field.md` with exactly this content**

```markdown
# Text field

A labelled input for one value. The label stays above the field and stays visible while the field has a value.

Source: docs/research/stripe-design-system.md section 2 (input label type), section 3 (input border), section 1 (focus ring).
```

- [ ] **Step 4: Write `components/empty-state.md` with exactly this content**

```markdown
# Empty state

What a list or page shows when it has no data yet: a title that says what is missing, a one-line description of when data appears, and one action.

Source: docs/research/stripe-design-system.md section 6 (Stripe Apps empty-state rules).
```

- [ ] **Step 5: Write `templates/payments-list.md` with exactly this content**

```markdown
# Payments list

A filterable, exportable list of payments. Each row shows the amount, the customer, the status and the date, and opens a detail page. When the list is empty it shows the empty state.

Source: docs/research/stripe-design-system.md section 7 (payments list in the Dashboard).
```

- [ ] **Step 6: Write `flows/account-setup.md` with exactly this content**

```markdown
# Account setup

The steps a new user takes to activate an account before going live:

1. Create an account.
2. Work in a sandbox right away.
3. Complete the account checklist.
4. Verify the business. This is required for live mode.
5. Set public details: name, URL, support contact and statement descriptor.
6. Secure the account with two-factor authentication and team invites.

Each step has one primary action. Live mode stays locked until step 4 is complete.

Source: docs/research/stripe-design-system.md section 7 (account setup and Dashboard onboarding).
```

- [ ] **Step 7: Run the pack tests**

Run: `go test ./internal/designsystem -run 'TestPack' -v`
Expected: `TestPackLoadsClean`, `TestPackShipsNoBrandAssets`, `TestPackFixtureHasNoFindings`, `TestPackSeededViolationIsFound` and `TestPackPatternsOnSnippets` PASS. `TestPackReadmeStatesUnofficial` still FAILS, because `packs/stripe-reference/README.md` is created in Task 5.

- [ ] **Step 8: Run the whole designsystem package**

Run: `go test ./internal/designsystem`
Expected: PASS once Task 5 is done. Any other failure must be read and fixed in this task's files, not by editing `pack_test.go`.

---

### Task 5: Pack README and root README pointer

**Files:**
- Create: `packs/stripe-reference/README.md`
- Modify: `README.md` (append a section at the end)

**Interfaces:**
- Consumes: `TestPackReadmeStatesUnofficial` (Task 1) requires the strings `unofficial`, `not endorsed by stripe` and `docs/research/stripe-design-system.md` in lower case.
- Produces: the pack README, read by the test.

**Model: haiku**

- [ ] **Step 1: Write `packs/stripe-reference/README.md` with exactly this content**

```markdown
# Stripe-inspired reference pack

**Unofficial.** This pack is a reverse engineering of Stripe's public design system, built only from publicly observable sources. It is not a Stripe product, it is not endorsed by Stripe, and it ships no Stripe logos, wordmarks, font files or other brand assets. Every part is labelled "inspired by".

It is a worked example: a complete design system (point of view, tokens, rules, components, a template and a flow) that `critique` can check.

## Sources

Every observation traces to [docs/research/stripe-design-system.md](../../docs/research/stripe-design-system.md). That file lists its public sources, the date accessed (2026-10-18), and a confidence level for each observation.

Token values are this pack's own choices, inspired by the hue families the research observed. No observed value is copied.

| Token category | Based on research section |
|---|---|
| color | 1. Colour palette and tokens |
| type | 2. Typography |
| space | 3. Spacing, radius, shadows, borders |
| radius | 3. Spacing, radius, shadows, borders |
| motion | 4. Motion |

Each item file cites its own sections in a `Source:` line.

## Check a page

Validate the pack from the repository root:

    go run ./cmd/tenet validate --dir packs/stripe-reference

Ask your MCP client to call `critique` with `file` set to a UI source file and `system` set to `packs/stripe-reference`. The fixture `testdata/critique/pack-clean.html` uses only pack tokens and reports no findings.

## Known limits

- Patterns check one line at a time, and check text, not structure.
- The spacing pattern flags raw pixel values even where the rule allows them.
- The motion pattern flags `ease-in-out` as well as `ease-in`, `ease-out` and `linear`.
- The judgement rules `button.one-primary` and `empty-state.title-states-missing` are not checked by `critique`; the client reads them through the delivery tools.
- Sail, Stripe's internal design system, and Dashboard visuals were not observed. This pack does not claim to match them.
```

- [ ] **Step 2: Append a pointer to the root `README.md`**

Read the end of `README.md` first, then append exactly this section after its last line, with one blank line before it:

```markdown
## Reference pack

A worked, unofficial Stripe-inspired design system lives in `packs/stripe-reference/`. See its [README](packs/stripe-reference/README.md).
```

- [ ] **Step 3: Run the designsystem tests**

Run: `go test ./internal/designsystem -v`
Expected: every `TestPack*` test PASS, and the rest of the package still PASS.

---

### Task 6: Server test for the pack fixture

**Files:**
- Modify: `internal/server/critique_test.go` (append one test)

**Interfaces:**
- Consumes: the existing `connect(t, root)` and `callTool(t, session, name, args)` helpers and `textOf` in `internal/server`.
- Produces: one test, `TestCritiquePackFixtureNoFindings`.

**Model: haiku**

- [ ] **Step 1: Append this test to `internal/server/critique_test.go`**

Add exactly this code at the end of the file:

```go

// TestCritiquePackFixtureNoFindings runs the reference pack against its fixture from the repository root.
// The pack lives outside testdata, so the server root is the repository root.
func TestCritiquePackFixtureNoFindings(t *testing.T) {
	session := connect(t, "../..")
	res := callTool(t, session, "critique", map[string]any{
		"file":   "testdata/critique/pack-clean.html",
		"system": "packs/stripe-reference",
	})
	want := "No findings in testdata/critique/pack-clean.html against packs/stripe-reference. Checked 4 rule(s).\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}
```

- [ ] **Step 2: Run the test**

Run: `go test ./internal/server -run TestCritiquePackFixtureNoFindings -v`
Expected: PASS. Every earlier task is in place, so this is the first run that can pass. If it fails, read the printed text; a count other than 4 means a rule's `Forbid:` line changed.

- [ ] **Step 3: Run the full suite**

Run: `go test ./...`
Expected: PASS for every package.

---

## Verification handed to slice-build (B5)

After Task 6 the slice-build flow runs the full checks itself: `go test ./...`, `test -z "$(gofmt -l .)"`, `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`, `go build -o tenet ./cmd/tenet`, and `go run ./cmd/tenet validate --dir packs/stripe-reference` (must exit 0). It also observes the critique result from the running server, as the acceptance criteria require.
