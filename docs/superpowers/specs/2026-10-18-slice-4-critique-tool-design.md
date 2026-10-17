# Slice 4: Critique tool — design

Approved: 2026-10-18

Self-approved under the project's no-approval-gates rule (see CLAUDE.md, "Decide, don't ask"). Path: architectural (new MCP tool, a new rule format, a new public type in `designsystem`).

## Goal

Expose a `critique` tool that checks one UI source file against the rules of a design system and returns one finding per breach. Each finding names the rule broken, the line, the text that broke it, and a specific fix. The server has no model access, so detection is deterministic: a rule that can be checked carries a pattern, and the tool reports where the pattern matches.

## Acceptance criteria (from docs/roadmap.md)

1. Tool `critique` takes a file path to a UI source file and the relevant system, and returns a list of findings, each naming the rule it breaks and a specific fix.
2. Findings reference rule IDs from the rules file.
3. A test fixture with two seeded violations reports both and no others.

## Decisions

| # | Decision | Reason |
|---|---|---|
| C1 | A rule may carry one `Forbid: <pattern>` body line and one `Fix: <text>` body line. The pattern is a Go RE2 regular expression. A match on a line of the source is a violation. | The server cannot judge intent; a pattern is the checkable part of a rule. RE2 is in the standard library and cannot backtrack catastrophically. |
| C2 | `Forbid:` and `Fix:` come as a pair. A `Forbid:` without `Fix:`, a `Fix:` without `Forbid:`, a second line of either kind, an empty pattern or an empty fix, or a pattern that does not compile is a problem at that line. The rule is then invalid (`Valid` false), as with any other parse problem (slice 3 D5). | A violation with no fix is not the specific fix acceptance 1 asks for. An invalid pattern must not silently check nothing. |
| C3 | `Forbid:` and `Fix:` are not part of the rule statement. They are recognised by their prefix at the start of a trimmed body line, like `Reason:` and `Applies to:`. | Keeps statements clean (slice 3 D2) and keeps the existing rules file format valid. |
| C4 | A rule without `Forbid:` is valid and is never reported by `critique`. It is still a rule for every other tool. | Many rules (for example point-of-view rules) cannot be checked by a pattern. Matches the spirit of slice 3 D4. |
| C5 | `critique` checks every valid rule that has a `Forbid:` line. An `Applies to:` line does not narrow the check. | The source file's component is not known, so the rule's scope cannot be decided from the file. Scope narrowing is a later decision, if a need appears. |
| C6 | `critique` matches each pattern against each line of the source, using `FindAllString`. Each match is one finding, so a line with two matches gives two findings. Matching is per line; a pattern cannot span lines. | Line numbers are exact and the result is easy to read. The limit (no multi-line patterns) is stated in the README. |
| C7 | Findings are ordered by line, then by rule order in `rules.md`, then by match order in the line. The order is deterministic. | Tests and reviewers can compare output exactly. |
| C8 | The tool has two required arguments: `file` (a path to the source file, relative to the project root or absolute) and `system` (a folder that holds its own `design-system/`, relative to the project root or absolute; `.` is the project root). | Acceptance 1 names both. A folder holding its own `design-system/` is the shape slice 5's `packs/stripe-reference/` needs, so `critique` runs against a pack without a new option. |
| C9 | `file` must exist, must be a regular file, must be at most 1 MiB, and must not contain a NUL byte in its first 8000 bytes. Otherwise the result is an `IsError` naming the file and the reason. | A UI source file is text, and an unbounded read of a huge or binary file is a foot-gun for an MCP server. |
| C10 | `system` must contain a `design-system/` folder. If the loader reports `design-system` itself as a problem (missing or not a directory), the result is that problem with the `run tenet init` text, as in slice 3 D12. | One root cause, one message. |
| C11 | If any problem names `design-system/rules.md`, the result ends with a blank line and `Note: design-system/rules.md has problems, so some rules may not be checked. Run tenet validate.` | A rule dropped by a parse problem must not be silent (slice 3 D14). |
| C12 | Output format. With findings: a header `<n> finding(s) in <file> against <system>.`, a blank line, then one line per finding: `<file>:<line>: <rule-id>: found "<match>". <fix>`. With no findings: `No findings in <file> against <system>. Checked <n> rule(s).` With zero checkable rules: `No rule in <system> has a Forbid: line, so nothing was checked. Add one under a rule in design-system/rules.md.` `<file>` is the path as the caller gave it. `<system>` is the path as the caller gave it. | Plain text, matching the other tools. Each finding line is self-contained, so the client can quote it. |
| C13 | `critique` calls `designsystem.Load` fresh on every call, as the delivery tools do (slice 3 D13). | Edits to the pack and the source show without a restart. |
| C14 | `critique` checks the file as text. It does not parse HTML, JSX or CSS. Patterns are written by the rule author against the text. | No parser dependency; the same mechanism serves any UI source language. |
| C15 | The bundled `testdata/project/design-system/rules.md` gains a `Forbid:` and a `Fix:` line on its two rules. The scaffold from `tenet init` is unchanged (its rules are optional to check). | The example shows the format. Scaffold stays valid by construction (slice 2 D26). |
| C16 | A new fixture `testdata/critique/page.html` holds exactly two seeded violations: a raw hex colour (`color.tokens-only`) and a raw pixel margin (`spacing.unit-multiples`). Its other lines use `var(--...)` and are clean. | Acceptance 3. The fixture is checked in and read by the test. |
| C17 | The spacing pattern approximates the rule. `spacing.unit-multiples` says spacing is a multiple of the space unit; the pattern flags any raw `px` value in `margin`, `padding` or `gap`. The fix names the space token instead. The spec records this as a known gap: a pixel value that is a multiple of the unit is also flagged. | RE2 cannot test divisibility. A stricter check is a later decision. The approximation is stated, not hidden. |
| C18 | `Rule` gains `Forbid *regexp.Regexp` (nil when the rule has none) and `Fix string`. `parseRules` sets them. `Valid` covers them (C2). | Lets the critique logic in `designsystem` use the parsed pattern without re-parsing. |
| C19 | The critique matching is a pure function in `designsystem`: it takes the source text and the rules, and returns findings. It does no file IO. The server reads the file and renders the result. | Keeps the package boundary from ADR-0007 and the slice 3 pattern (D15). The matching is testable without a server. |
| C20 | ADR-0009 records the decision that rules carry an optional deterministic pattern and that judgement (for example whether a layout fits the user journey) stays with the client, which reads the rules through the delivery tools. | The format change and the split between server and client are architectural and need a record. |
| C21 | `GLOSSARY.md` gains **Finding**: one breach of one rule in one file, with its line and fix. | Terms used in code and output stay in the glossary. |
| C22 | The README gains: the `Forbid:` and `Fix:` lines, the `critique` tool, its arguments, its output, and the line-only limit (C6). | The README documents the public surface (slice 3 did the same for `Applies to:`). |

## Architecture

```
internal/designsystem/
├── rules.go           # parseRules: Forbid:/Fix: lines, compile the pattern, Valid
├── rules_test.go      # extended: Forbid/Fix cases
├── critique.go        # Finding type; Critique(source string, rules []Rule) []Finding (pure)
├── critique_test.go   # matching, ordering, invalid rules skipped
└── ... (unchanged)
internal/server/
├── server.go          # New registers the critique tool
├── critique.go        # CritiqueInput, handler: file checks, Load, Critique, rendering
├── critique_test.go   # in-memory client tests: fixture findings, error paths, notes
└── server_test.go     # tool-list test updated to six tools
testdata/critique/page.html         # the fixture (C16)
testdata/project/design-system/rules.md   # Forbid:/Fix: lines (C15)
docs/adr/0009-rules-carry-deterministic-checks.md
GLOSSARY.md                          # Finding (C21)
README.md                            # C22
```

Dependency direction is unchanged: `cmd` → `server` → `designsystem`. `designsystem` imports only the standard library and has no MCP types.

## Module interfaces

**`internal/designsystem`**

```go
// Rule gains two fields.
type Rule struct {
    // ... existing fields ...
    Forbid *regexp.Regexp // nil when the rule has no Forbid: line
    Fix    string         // text of the Fix: line; empty when none
}

// Finding is one breach of one rule in one source file.
type Finding struct {
    Line   int    // 1-based line in the source
    RuleID string
    Match  string // the text the pattern matched
    Fix    string // the rule's Fix: text
}

// Critique returns the findings of the valid rules that have a Forbid pattern, in the order of C7.
func Critique(source string, rules []Rule) []Finding

// Checkable returns the number of valid rules that have a Forbid pattern.
func Checkable(rules []Rule) int
```

**`internal/server`**

```go
// CritiqueInput is the input of critique.
type CritiqueInput struct {
    File   string `json:"file" jsonschema:"path to the UI source file, relative to the project root or absolute"`
    System string `json:"system" jsonschema:"folder holding the design-system/ to check against, relative to the project root; . for the project root"`
}
```

The handler returns `*mcp.CallToolResult`; all failures are `IsError` results (slice 1 convention).

## Data flow

1. Resolve `file` against the project root when relative. `os.Stat`: missing, a directory, or over 1 MiB → `IsError`. Read the bytes. NUL in the first 8000 bytes → `IsError` ("is not a text file").
2. Resolve `system` against the project root when relative. `designsystem.Load(system)`.
3. If a problem names `design-system` itself → `IsError` with the `run tenet init` text (C10).
4. `Critique(source, pack.Rules)` gives the findings. `Checkable(pack.Rules)` gives the count.
5. Render per C12. If a problem names `design-system/rules.md`, append the note (C11).

## Errors

| Situation | Result |
|---|---|
| `file` missing | `file "<file>" not found.` |
| `file` is a directory | `file "<file>" is a directory; pass a file.` |
| `file` over 1 MiB | `file "<file>" is larger than 1 MiB; pass a smaller source file.` |
| `file` contains NUL | `file "<file>" is not a text file.` |
| `system` has no `design-system/` | `design-system not found under <system>; run tenet init there.` (via C10) |
| `system`'s design-system is invalid at `design-system` | the problem text plus `; run tenet init` (C10) |

A pattern that does not compile is not an error of `critique`: the rule is invalid, `validate` reports it, and the rule is not checked (C2).

## Testing

- `rules_test.go`: `Forbid:`/`Fix:` cases: a valid pair; `Forbid:` alone; `Fix:` alone; two `Forbid:` lines; empty pattern; empty fix; an uncompilable pattern (problem at the line, rule invalid); a rule without either line stays valid with `Forbid` nil.
- `critique_test.go` (designsystem): a line with two matches gives two findings; findings sort by line then rule order; an invalid rule is skipped; a rule with no `Forbid` is skipped; `Checkable` counts valid rules with a pattern.
- `critique_test.go` (server): the fixture against `testdata/project` returns exactly two findings, one for `color.tokens-only` at line 4 and one for `spacing.unit-multiples` at line 5, with the rule IDs, matches and fixes; no other lines; the file-error cases in the Errors table; a system with no checkable rules gives the no-rule message; a rules file with a problem adds the note.
- `server_test.go`: the tool list has six tools, sorted by name.
- Full suite: `go test ./...`, `gofmt`, golangci-lint (the commands in CLAUDE.md).

## Known limits (stated in the README)

- Patterns match one line at a time (C6).
- A pattern checks text, not structure (C14).
- The spacing pattern flags raw pixel values in spacing properties even when they are multiples of the unit (C17).
- The tool reports only what a `Forbid:` pattern can express. Judgement against the point of view and the user journey stays with the client, which reads the rules through the delivery tools (ADR-0009).
