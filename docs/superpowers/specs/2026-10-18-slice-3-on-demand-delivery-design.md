# Slice 3: On-demand delivery tools — design

Approved: 2026-10-18

Self-approved under the project's no-approval-gates rule (see CLAUDE.md, "Decide, don't ask"). Path: architectural (new MCP tools, a new cross-reference format between rules and items, changes to the loader's public types).

## Goal

Expose the loaded design system to MCP clients through four tools, `get_tokens`, `get_component`, `get_template` and `get_flow`. Each returns only the requested item and the rules linked to it, so one call stays small. Each error names what was missing and lists what exists.

## Acceptance criteria (from docs/roadmap.md)

1. Tools `get_tokens`, `get_component`, `get_template`, `get_flow` return only the requested item and its linked rules, so a single call stays small (under 4 KB for the bundled example).
2. Each tool returns an error that names the missing item and lists the available ones.

## Decisions

| # | Decision | Reason |
|---|---|---|
| D1 | Linking is written in `rules.md`, not in item files. A rule entry may have one optional body line `Applies to: <ref>, <ref>, ...`. Each ref is `kind:id`, where kind is `component`, `template`, `flow` or `tokens`. For `tokens`, the id is a category name. | Item files keep the format from slice 2 (a title and free Markdown). Rules are already the unit critique will cite, so the rule owns its scope. |
| D2 | The `Applies to:` line is not part of the rule statement. A rule has at most one such line. An empty line, an unknown kind, a missing colon, an invalid id, or an empty entry (for example a trailing comma) is a problem at that line. The entry is skipped. | Keeps statements clean; makes a bad link visible rather than silently unlinked. |
| D3 | A ref to an item or category that does not exist is a problem reported by `Load` at the `Applies to:` line. The rule itself stays valid. | Cross-reference checks were deferred by slice 2 D17 to the first slice that uses references. This is that slice. A missing target is a rule-authoring mistake, not a rule defect. |
| D4 | A rule with no `Applies to:` line is valid and is never returned by an item or token tool. Rules are global unless linked. | Global rules (for example the point of view's own rules) are delivered by no tool yet; avoids guessing a scope. |
| D5 | `Rule` gains `AppliesTo []Ref` and `Valid bool`. `Valid` is false when the entry produced any problem in `parseRules`, including a missing or empty `Reason:` line. Delivery serves only valid rules, in `rules.md` order. | A rule without a reason must not reach the AI as if it were complete. Problems in cross-references (D3) do not change `Valid`. |
| D6 | `Ref` is `{Kind, ID string; Line int}`. `Kind` is one of the constants `KindComponent`, `KindTemplate`, `KindFlow`, `KindTokens`. | Gives the loader and the lookups one vocabulary, and lets `Load` point at the exact line. |
| D7 | `Item` gains `Body string`: the full file text as read. | The item tools return the file verbatim, so the AI sees exactly what the author wrote. |
| D8 | `get_component`, `get_template`, `get_flow` take one required argument, `id` (string). The result is the item's `Body` with trailing whitespace trimmed, then, only if linked valid rules exist, a blank line, `## Linked rules`, and for each rule a blank line, `### <id>`, the statement on its own line, and `Reason: <reason>`. | Matches the format authors already write in `rules.md`, so the AI reads both the same way. |
| D9 | `get_tokens` takes one required argument, `category` (string). The result is one line per token of that category in JSON order, `category.name: value`, then the linked rules in the same format as D8. | Tokens are values, not prose; one line per token is the smallest faithful form. A category is the natural unit a component's tokens belong to (glossary: Delivery). |
| D10 | Errors are `IsError` tool results, not Go errors, as in slice 1. Not-found text is `<kind> "<id>" not found. Available <kinds>: a, b.` with ids sorted, or `<kind> "<id>" not found. The design system has no <kinds> yet; add Markdown files under design-system/<dir>/.` when there are none. For tokens, the kind is `token category` and the directory is `tokens.json`. | Acceptance 2. Sorted lists make the output deterministic for tests. |
| D11 | If the requested item's file is in the loader's problems, the error is `<kind> "<id>" is not valid: <problem>. Run tenet validate.` (for the first problem with that file). It is not reported as not found. | Otherwise a broken file looks like a typo and the list of available items would not show why. |
| D12 | If the loader reports `design-system` itself as a problem (missing or not a directory), every tool returns that problem with `run tenet init` text, and nothing else. | One root cause, one message (slice 2 D24). |
| D13 | Each tool call calls `designsystem.Load(root)` fresh. `server.New(root)` keeps its signature. | Edits to the pack show without restarting the server, as slice 1's point-of-view tool already does. This supersedes slice 2 D20, which expected `server.New` to receive the loaded `DesignSystem`; the server never caches the pack, so there is nothing to pass. |
| D14 | If any problem has `File == design-system/rules.md`, every delivery response ends with a blank line and `Note: design-system/rules.md has problems, so some linked rules may be missing. Run tenet validate.` | A missing rule must not be silent. Only the rules file is named because only it can drop a linked rule. |
| D15 | Lookup functions live in `internal/designsystem` as pure methods on `DesignSystem`, with no MCP types. Rendering and tool registration live in `internal/server`. | Keeps the package boundary from ADR-0007 and the slice 2 layout. The lookups are testable without a server. |
| D16 | `get_point_of_view` is unchanged. The server's tool list is the five tools, sorted by name in tests. | Acceptance criteria only add tools; slice 1 behaviour is kept. |
| D17 | Only the fixture gains `Applies to:` lines: `color.tokens-only` applies to `tokens:color` and `component:button`; `spacing.unit-multiples` applies to `tokens:space` and `template:settings-page`. The flow `first-run` has no linked rules, on purpose, so the no-rules path is exercised. | Gives the server tests and the README a real linked example. |
| D18 | The `design-system/` scaffold from `tenet init` is unchanged. Its `rules.md` preamble still describes the entry format; the README documents `Applies to:`. | Scaffold must stay valid by construction (slice 2 D26); `Applies to:` is optional. |
| D19 | No new subcommand, flag or config. `validate` reports the new `Applies to:` problems through the existing loader path. | Nothing new to learn at the command line. |

## Architecture

```
internal/designsystem/
├── load.go            # Item.Body; Rule.AppliesTo/Valid; cross-reference check after parsing
├── rules.go           # parseRules: Applies to: line, Ref parsing, Valid flag
├── items.go           # loadItems: fills Item.Body
├── lookup.go          # DesignSystem methods: Category, Item, Categories, Linked (pure, no MCP)
├── lookup_test.go
├── rules_test.go      # extended: Applies to: cases
├── load_test.go       # extended: cross-reference problems, Body
└── ... (unchanged files from slice 2)
internal/server/
├── server.go          # New registers the four delivery tools; get_point_of_view unchanged
├── delivery.go        # tool input types, handlers (Load per call), rendering, not-found and invalid errors
├── delivery_test.go   # in-memory client tests for the four tools
└── server_test.go     # tool-list test updated to five tools
testdata/project/design-system/rules.md   # Applies to: lines added (D17)
README.md                                  # Applies to:, four tools, the 4 KB note
```

Dependency direction is unchanged: `cmd` → `server` → `designsystem`, and `cmd` → `designsystem`. `designsystem` imports only the standard library and has no MCP types.

## Module interfaces

**`internal/designsystem`**

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

// Rule gains two fields.
type Rule struct {
    ID        string
    Statement string
    Reason    string
    Line      int
    AppliesTo []Ref // in the order written
    Valid     bool  // false when parseRules reported any problem for this entry
}

// Item gains Body.
type Item struct {
    ID    string
    Title string
    Path  string
    Body  string // full file text as read
}

// Category returns the tokens of one category, in file order, and whether the category exists.
func (ds DesignSystem) Category(name string) ([]Token, bool)

// Categories returns the category names in sorted order, without duplicates.
func (ds DesignSystem) Categories() []string

// Item returns the item of the given kind (KindComponent, KindTemplate or KindFlow) with the given id.
func (ds DesignSystem) Item(kind Kind, id string) (Item, bool)

// Items returns the ids of the items of the given kind, sorted.
func (ds DesignSystem) ItemIDs(kind Kind) []string

// Linked returns the valid rules whose AppliesTo contains kind:id, in rules.md order.
func (ds DesignSystem) Linked(kind Kind, id string) []Rule
```

`Load` keeps its signature. Its problems gain the new `Applies to:` problems, and `Rule.Valid` is set by `parseRules`.

**`internal/server`**

```go
// GetTokensInput is the input of get_tokens.
type GetTokensInput struct {
    Category string `json:"category" jsonschema:"the token category, for example color"`
}

// GetItemInput is the input of get_component, get_template and get_flow.
type GetItemInput struct {
    ID string `json:"id" jsonschema:"the item id, the file name without .md"`
}

// renderItem and renderTokens are unexported; they build the text of a delivery result.
// deliveryError is unexported; it builds the IsError result for not-found and invalid items.
```

The handlers take the root, call `designsystem.Load(root)`, and return `*mcp.CallToolResult` with `IsError` as needed. Required inputs follow go-sdk v1.8.0's schema inference; the plan confirms that a missing `category` or `id` is a protocol-level argument error, not a tool result.

## Data flow

**get_component / get_template / get_flow**
1. `Load(root)`.
2. If problems name `design-system` itself, return that problem with the `tenet init` text (D12).
3. `Item(kind, id)`. Found: render the body and `Linked(kind, id)` (D8), plus the rules-file note if needed (D14).
4. Not found: if a problem names the item's file, return the invalid error (D11). Otherwise return the not-found error with `ItemIDs(kind)` (D10).

**get_tokens**
1. `Load(root)`, then the same `design-system` check.
2. `Category(name)`. Found: one line per token, then `Linked(KindTokens, name)` (D9), plus the note if needed.
3. Not found: if the tokens file has a problem, return it as the invalid error. Otherwise return the not-found error with `Categories()`.

**Load, cross-references**
After all files are parsed, each `Ref` in each rule is checked: component, template and flow refs against the item lists, tokens refs against the token categories. An unknown target is a problem at the ref's line (D3). `parseRules` runs first and sets `Valid`, so cross-reference problems never change it.

## Error and edge cases

| Case | Behaviour |
|---|---|
| `design-system/` missing | Every delivery tool: `design-system: directory not found; run tenet init`, as an IsError result. |
| `design-system` is a file | Same, with `not a directory`. |
| Item id not in pack, others exist | Error `component "x" not found. Available components: button, card.` |
| Item id not in pack, none exist | Error `component "x" not found. The design system has no components yet; add Markdown files under design-system/components/.` |
| Item id exists but its file has a problem | Error `component "x" is not valid: <problem>. Run tenet validate.` |
| Id is empty | Treated as an unknown id; the not-found error lists available items. |
| Category not in pack, others exist | Error `token category "x" not found. Available token categories: color, space.` |
| No token categories | Error `token category "x" not found. The design system has no tokens yet; add categories to design-system/tokens.json.` |
| Tokens file has problems and the category is missing | Error with the first problem for `design-system/tokens.json`, as the invalid error. |
| Item with no linked valid rules | Body only, no `Linked rules` section. |
| Linked rule is invalid | Omitted from the response. The rules-file note appears (D14). |
| Linked rule points at a missing target | The rule still counts for the targets that exist; the missing one is a Load problem only. |
| Rule with `Applies to:` but invalid `Reason:` | Not delivered (`Valid` false). The rules-file note appears. |
| Two `Applies to:` lines in one entry | Problem at the second line; the second line's refs are dropped; the first line's refs stand. |
| Empty `Applies to:` text | Problem at that line: `rule "x" has an empty Applies to: line`. |
| Bad ref kind, such as `thing:x` | Problem at the line: `rule "x" has an unknown kind "thing" in Applies to:; use component, template, flow or tokens`. The entry is kept with the valid refs. |
| Ref without a colon | Problem: `rule "x" has an Applies to: entry "y" without a kind; write kind:id`. |
| Ref with an invalid id | Problem: `rule "x" has an invalid id "Y" in Applies to:`. |
| Ref to a missing item or category | Problem at the ref line: `rule "x" applies to unknown component "button"` (or `token category`). |
| Edit to a file while the server runs | The next call sees it, because each call loads. |

## Testing

Test-first for every behaviour. Each row below is one test case in a table, unless noted.

**`internal/designsystem`**
- `TestParseRules` extended: a valid `Applies to:` with two refs; each kind; each error row above with its line; repeated and empty lines; `Applies to:` inside a fence ignored; statement excludes the `Applies to:` line; `Valid` true for a clean entry and false for each entry with a problem (no reason, empty reason, repeated reason, bad id, no statement).
- `TestLoadReferences`: a rule linked to an existing component, token category, template and flow yields no problems; a rule linked to a missing component or category yields one problem at the ref's line; the rule remains in `Rules` with `Valid` true.
- `TestLoadItemsBody`: `Body` equals the file content (BOM stripped).
- `TestLookups`: `Category` returns tokens in file order and reports missing categories; `Categories` sorted and unique; `Item` finds by kind and id and does not confuse kinds; `ItemIDs` sorted; `Linked` filters on kind and id, skips invalid rules, and keeps rules.md order; a rule linked to two targets appears under each.
- `TestFixtureLinks`: `Linked(KindComponent, "button")` returns `color.tokens-only` only; `Linked(KindFlow, "first-run")` is empty; `Category("color")` has three tokens.

**`internal/server`**
- `TestListsFiveTools`: the list is `get_component`, `get_flow`, `get_point_of_view`, `get_template`, `get_tokens`, sorted, and each input schema is an object with the expected required field.
- `TestGetComponentFixture`: text starts with the button file body and contains `## Linked rules`, `### color.tokens-only`, and its `Reason:` line; IsError false.
- `TestGetFlowFixtureHasNoLinkedRules`: body only, no `Linked rules`.
- `TestGetTokensFixture`: contains `color.ink: #1d2433` and does not contain `space.unit`; contains the linked rule.
- `TestNotFoundListsAvailable`: `get_component` with `id: "card"` on the fixture returns IsError with `Available components: button.`; the same for template and flow, and for `get_tokens` with `category: "motion"` (`Available token categories: color, space.`).
- `TestNotFoundWhenKindEmpty`: a temp root with a tokens file and no items returns the `has no components yet` text.
- `TestInvalidItemReportsProblem`: a component file whose first line is not a heading returns the invalid error with the problem and `run tenet validate`.
- `TestMissingDesignSystemIsToolError`: an empty temp root returns IsError containing `run tenet init` for every tool.
- `TestEditsVisibleWithoutRestart`: connect, call, rewrite the button file, call again, and the new text is returned.
- `TestRulesProblemNote`: a temp root whose rules file has an invalid rule linked to a component: the response ends with the note.
- `TestResultsUnder4KB`: each of the four tools on the fixture (with valid ids) returns fewer than 4096 bytes of text.

**Quality gates**, all of which must pass before the slice is committed:
- `go test ./...`
- `test -z "$(gofmt -l .)"`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`
- `go run ./cmd/tenet validate --dir testdata/project` exits 0
- Stdio check from the README returns the point of view and one delivery tool result.

## Repository files

- `testdata/project/design-system/rules.md`: `Applies to:` lines per D17.
- `README.md`: the `Applies to:` line in the rules section (its format, the kinds, and that rules without it are not delivered by any tool), a section on the four delivery tools with their arguments and the 4 KB note, and the stdio check gaining one delivery call. The existing `validate`, `init` and Claude Code sections stay.
- `CLAUDE.md`: unchanged here. Slice-build's closing step proposes any project-section additions.

## Out of scope for slice 3

Critique and its findings (slice 4). Items linking to rules from the item side. Global rules delivered by any tool. Search or listing across kinds in one call. Filtering tokens by name, pagination, or size limits on item bodies beyond the 4 KB example check. Nested folders in item directories. Structured (JSON) token output. The Stripe reference pack (slice 5).

## Open risks

- Authors must remember to add `Applies to:` to a rule for it to be delivered. A rule written without it is valid and silently unlinked, so it never reaches an item call. The README states this; a later slice can add a "rules with no scope" validation warning if it proves to bite.
- Item bodies are returned verbatim with no size limit. A large component file can exceed 4 KB; the bundled example does not. A later slice should decide on a limit once real packs exist.
- Each call re-reads the whole design system. This is cheap for a hand-written pack, but a large pack with many items may need caching; that is a later decision, made with measured cost.
- The rules-file note (D14) names the file, not the broken rule, so an author must run `tenet validate` to see which rule dropped out.
