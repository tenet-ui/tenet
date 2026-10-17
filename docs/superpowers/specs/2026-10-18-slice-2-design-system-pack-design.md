# Slice 2: Design system pack format and validation — design

Approved: 2026-10-18

Self-approved under the project's no-approval-gates rule (see CLAUDE.md, "Decide, don't ask"). Path: architectural (new subsystem, new file formats, new CLI subcommands).

## Goal

Define the on-disk format of a `design-system/` folder (point of view, tokens, rules, components, templates, flows), load it through one shared code path, check it with `tenet validate`, and scaffold a valid empty one with `tenet init`. Nothing exposes the loaded pack to MCP clients yet; that is slice 3.

## Acceptance criteria (from docs/roadmap.md)

1. A `design-system/` folder with tokens (JSON), rules (Markdown, one rule per entry with a reason), components, templates and flows is read by the server.
2. A `tenet validate` command reports missing or malformed files with file and line, and exits non-zero on error.
3. A `tenet init` command scaffolds an empty, valid `design-system/` folder.

## Decisions

| # | Decision | Reason |
|---|---|---|
| D1 | Subcommands: no subcommand serves (unchanged from slice 1). `tenet validate [--dir <path>]` and `tenet init [--dir <path>]`. `--dir` keeps its slice 1 meaning and default. | Keeps the existing `tenet --dir` invocation in the README and in users' `claude mcp add` lines working. |
| D2 | An unknown first argument that is not a flag (`tenet foo`) exits 2 with `unknown command "foo"` on stderr. | Otherwise `foo` would be silently ignored and the server would start. |
| D3 | `design-system/point-of-view.md` is required and must not be whitespace-only. Its format is free Markdown. | Slice 1 rule (D11) carried forward. |
| D4 | `design-system/tokens.json` is required. Top level is an object whose keys are categories; each category is an object whose keys are token names; each token value is a JSON string or number. | Two-level JSON is the smallest shape that names a token `category.name`, which slice 3 and critique will reference. |
| D5 | Category and token names match `^[a-z][a-z0-9-]*$`. Categories are not restricted to the glossary set (colour, space, radius, type, motion); any valid name is accepted. | Glossary lists examples, not a closed set. Restricting now would force a format change the first time a team needs `elevation` or `z-index`. |
| D6 | Token value: a non-empty string, or a number (kept as its literal text). Bools, null, arrays, and nested objects are errors. | Tokens are design values; structured tokens are out of scope until a slice needs them. |
| D7 | Duplicate category keys or duplicate token names within a category are errors, reported at the second occurrence. | `encoding/json` silently keeps the last duplicate, which hides a real mistake. |
| D8 | Trailing content after the top-level object, and an empty file, are errors. | Catches a half-saved or concatenated file. |
| D9 | `design-system/rules.md` is required. Zero entries is valid. Preamble text before the first entry is free text. | Lets `tenet init` produce a valid empty system (acceptance 3). |
| D10 | A rule entry starts at a line `## <id>` outside a fenced code block. The id matches `^[a-z0-9]+([.-][a-z0-9]+)*$` (for example `color.tokens-only`). Ids are unique within the file. | Ids are what critique will cite (roadmap slice 4). Dotted and dashed slugs match the token naming in D5 and read well in findings. |
| D11 | Each rule entry needs (a) a statement: at least one non-blank body line that is not the `Reason:` line; (b) exactly one `Reason:` line with non-empty text. Missing or empty parts are errors, reported at the heading line (or at the `Reason:` line when it is empty or repeated). | ADR-0002 and the roadmap require one rule per entry with a reason. A rule without a reason cannot be checked for intent. |
| D12 | Fenced code blocks (lines starting with ` ``` `) are tracked, so a `## ` inside a fence is not a rule. | Rules may quote example Markdown; without this, quoted headings would become bogus rules. |
| D13 | `design-system/components/`, `templates/`, `flows/` are optional directories. A missing directory means zero items, not an error. If one of these paths is a regular file, that is an error. | Teams start with tokens and rules and add the rest later. Scaffolding still creates the folders so their place is visible. |
| D14 | Each `*.md` file directly inside a component, template, or flow folder is one item. Non-`.md` entries and subdirectories are ignored. | Keeps `.gitkeep` and editor leftovers from becoming items. Nested folders are not part of the format yet. |
| D15 | Item id is the file name without `.md`. It must match `^[a-z0-9]+(-[a-z0-9]+)*$`; otherwise the file is reported and left out of the pack. | Ids are URL- and tool-argument-friendly for slice 3 (`get_component` takes an id). |
| D16 | Item validity: the first non-blank line is a level-1 heading `# <title>` with non-empty title. An empty file is an error at line 1. | Gives every item a title for the summary and for slice 3's "available items" error list, with no other schema. |
| D17 | Markdown checks stay minimal: no cross-reference checks, no link checks, no front matter, no required sections. | No references exist in slice 2. Each additional check is a format decision that a later slice should make with real usage. |
| D18 | One loader: `designsystem.Load(root) (Pack, []Problem)` is the only code that reads the pack. `tenet validate` prints its problems. The serve path calls `Load` at startup and logs the summary and problems to stderr, then serves as before. | One definition of "valid" for both commands and the server, so they cannot disagree. |
| D19 | A bad pack does not stop the server. `get_point_of_view` behaves exactly as in slice 1, reading the point-of-view file on each call. | Slice 1 D10 and D12. A broken token file should not take down the one tool that works. |
| D20 | `server.New(root)` keeps its signature in this slice. The loaded `Pack` is not passed to the server yet. | No tool uses the pack until slice 3. Slice 3 changes `server.New` to take the `Pack` when it needs it. This is the one place the server package does not "read" the pack directly; the serve path reads it on the server's behalf. |
| D21 | Problems are printed as `path:line: message`, or `path: message` when there is no line. Paths are slash-separated and relative to the project root. | The `file:line:` shape is what editors and terminals turn into clickable links. |
| D22 | `validate` prints problems to stdout, then either `design system valid: N tokens, M rules, C components, T templates, F flows` (exit 0) or `K problem(s) found` (exit 1). Usage and `--dir` errors exit 2. | Stdout is fine here: `validate` is not an MCP session. Exit codes follow slice 1's `--dir` error (2). |
| D23 | Line numbers are 1-based. For JSON, a problem is reported on the line where the offending token ends (for a nested object value, the line of its opening `{`). For a JSON syntax error, the line of the byte where the decoder failed. | `json.Decoder` gives byte offsets, not lines. These rules are exact for one-line tokens, which is how the format is written. |
| D24 | A missing `design-system/` directory is one problem, `design-system: directory not found; run tenet init`, and nothing else is checked. | Avoids five redundant "not found" lines for one root cause. |
| D25 | `tenet init` refuses if `design-system` exists in any form (file, directory, symlink) and exits 1 without writing anything. | Never overwrite a user's work. |
| D26 | `tenet init` writes: `point-of-view.md` (a `# Point of view` heading and a one-paragraph prompt to replace it), `tokens.json` (`{}`), `rules.md` (a `# Rules` heading and a one-line note on the entry format), and empty `.gitkeep` files in `components/`, `templates/`, `flows/`. It uses exclusive-create writes. If a write fails after the directory was created, it removes the directory it created and returns the error. | Scaffold is valid by construction (tested). Exclusive create avoids racing a concurrent writer. |
| D27 | Fixture `testdata/project` gets a complete, valid pack: `tokens.json` (a few colours and one space unit), `rules.md` with two rules, one component, one template, one flow. | Gives the server tests and the README a real example that passes `validate`. |
| D28 | Items and tokens are stored in the pack in file order (tokens in JSON order; rules in heading order; item folders sorted by file name). | Deterministic output for tests and for the summary. |

## Architecture

```
cmd/tenet/
├── main.go            # run(): subcommand dispatch; serve path; parseRoot (unchanged)
├── validate.go        # runValidate(): Load, print problems, exit code
├── init.go            # runInit(): designsystem.Scaffold, print result, exit code
└── main_test.go       # parseRoot tests (unchanged) + dispatch and command tests
internal/designsystem/
├── pointofview.go     # ReadPointOfView (unchanged from slice 1)
├── pointofview_test.go
├── load.go            # Load: orchestrates the file checks, returns Pack and Problems
├── tokens.go          # parseTokens: JSON token walk with line numbers
├── rules.go           # parseRules: Markdown rule entries
├── items.go           # loadItems: component/template/flow folders
├── problem.go         # Problem type and formatting
├── scaffold.go        # Scaffold: writes the empty design system
└── *_test.go          # table tests per file
internal/server/       # unchanged in this slice
testdata/project/design-system/
├── point-of-view.md
├── tokens.json
├── rules.md
├── components/button.md
├── templates/settings-page.md
└── flows/first-run.md
```

Dependency direction is unchanged: `cmd` → `server` → `designsystem`, and `cmd` → `designsystem`. `designsystem` imports only the standard library and has no MCP types.

## Module interfaces

**`internal/designsystem`**

```go
package designsystem

// Problem is one validation finding. Line is 1-based; 0 means the whole file.
type Problem struct {
    File    string // slash-separated, relative to the project root
    Line    int
    Message string
}

// String formats the problem as "file:line: message", or "file: message" when Line is 0.
func (p Problem) String() string

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
    Line      int    // line of the "## <id>" heading
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
func Load(root string) (Pack, []Problem)

var ErrAlreadyExists = errors.New("design-system already exists")

// Scaffold writes an empty, valid design-system/ folder under root. It returns
// an error wrapping ErrAlreadyExists, without writing anything, when the path
// exists in any form.
func Scaffold(root string) error
```

Unexported helpers the tests exercise directly: `parseTokens(file string, data []byte) ([]Token, []Problem)`, `parseRules(file string, data []byte) ([]Rule, []Problem)`, `loadItems(root, dir string) ([]Item, []Problem)`.

**`cmd/tenet`**

```go
// run dispatches on the first argument and returns the process exit code.
func run(args []string, cwd string, stdout, stderr io.Writer) int

// runValidate implements `tenet validate`. Exit: 0 valid, 1 problems, 2 usage or --dir error.
func runValidate(args []string, cwd string, stdout, stderr io.Writer) int

// runInit implements `tenet init`. Exit: 0 written, 1 already exists or write failed, 2 usage or --dir error.
func runInit(args []string, cwd string, stdout, stderr io.Writer) int

// parseRoot is unchanged from slice 1.
func parseRoot(args []string, cwd string) (string, error)
```

`run` takes `stdout` now (slice 1 had only `stderr`). Serve path: `run` with no subcommand calls `designsystem.Load(root)`, logs the summary and each problem to stderr, then `server.New(root).Run(ctx, &mcp.StdioTransport{})`.

Summary line on stderr at startup: `tenet <Version> serving <root> on stdio; design system: N tokens, M rules, C components, T templates, F flows`. If problems exist, one more line: `design system has K problem(s); run tenet validate for details`, then each problem as `tenet: <problem>`.

## Data flow

**Load:** `Load(root)`
1. `design-system` missing → one problem (D24), return. A regular file at that path → problem `design-system: not a directory`, return.
2. `point-of-view.md`: read with `ReadPointOfView`-style logic; missing, directory, or whitespace-only → problem (line 1 for empty).
3. `tokens.json`: `parseTokens`.
4. `rules.md`: `parseRules`.
5. `components/`, `templates/`, `flows/`: `loadItems` for each.
6. Return the pack and all problems, in the order above.

**validate:** `parseRoot` → `Load` → print problems → print summary or count → exit code.

**serve:** `parseRoot` → `Load` → log → `server.New` → `Run`.

**init:** `parseRoot` → `Scaffold` → print `created design-system/ under <root>` → exit 0. On `ErrAlreadyExists`, print the message to stderr and exit 1.

## Error and edge cases

| Case | Behaviour |
|---|---|
| No `design-system/` | One problem, `design-system: directory not found; run tenet init`. Validate exits 1. Serve logs it and still serves. |
| `design-system` is a file | One problem, `design-system: not a directory`. |
| `point-of-view.md` missing | Problem `design-system/point-of-view.md: file not found`. |
| `point-of-view.md` whitespace only | Problem at line 1: `file is empty`. |
| `point-of-view.md` is a directory | Problem `could not read: <err>`. |
| `tokens.json` missing | Problem `file not found`. |
| `tokens.json` empty (zero bytes) | Problem at line 1: `file is empty; write {} for no tokens`. |
| `tokens.json` invalid JSON | Problem at the failing line: `invalid JSON: <decoder message>`. Parsing stops; later tokens are not reported. |
| Top level is not an object | Problem at line of the first token: `top level must be an object of categories`. |
| Category is not an object | Problem at the category's key line: `category "<name>" must be an object of tokens`. Its tokens are skipped. |
| Invalid category or token name | Problem at the key line; that entry is skipped. |
| Token value is bool, null, array, or object | Problem at the key line: `token "<cat>.<name>" must be a string or number`. Value is skipped. |
| Token value is `""` | Problem at the key line: `token "<cat>.<name>" is empty`. |
| Duplicate category or token | Problem at the second key's line: `duplicate ...`. The first is kept. |
| Trailing content after the object | Problem at its line: `unexpected content after the top-level object`. |
| `rules.md` missing | Problem `file not found`. |
| `rules.md` with no entries | Valid. Zero rules. |
| Rule id invalid | Problem at heading line; the entry is skipped. |
| Rule id duplicate | Problem at second heading line: `duplicate rule id "<id>" (first at line N)`. The second entry is skipped. |
| Rule with no statement | Problem at heading line: `rule "<id>" has no statement`. |
| Rule with no `Reason:` line | Problem at heading line: `rule "<id>" has no Reason: line`. The rule is kept with an empty reason and fails validation. |
| `Reason:` with empty text | Problem at that line: `Reason: is empty`. |
| Second `Reason:` line in one entry | Problem at that line: `rule "<id>" has more than one Reason: line`. |
| `## ` inside a fenced code block | Not a heading; part of the preceding text or entry body. |
| Unclosed fence | Everything after the fence is inside it, so later rule headings are not seen. Not reported in this slice (D17). |
| `components/` missing | Zero components, no problem. |
| `components` is a file | Problem `components: not a directory`. |
| Non-`.md` file or subdirectory in an item folder | Ignored. |
| Item file name not kebab-case | Problem with line 0: `file name must be lowercase kebab-case, for example "primary-button.md"`. The item is left out of the pack. |
| Item file empty or whitespace only | Problem at line 1: `file is empty; start with a "# <title>" line`. |
| Item first non-blank line is not `# <title>` | Problem at that line: `first line must be a "# <title>" heading`. |
| Item `# ` heading with empty title | Same as above. |
| Scaffold target exists as file, directory, or symlink | `ErrAlreadyExists`; nothing written; exit 1. |
| Scaffold write fails midway | The directory created by this call is removed; error returned; exit 1. |
| Run with `--dir` missing or a file | Exit 2, as in slice 1. |
| Unknown subcommand | Exit 2, `unknown command "<name>"`. |

## Testing

Test-first for every behaviour. Each row below is one test case in a table, unless noted.

**`internal/designsystem`**
- `TestLoadFixtureIsValid`: `Load("../../testdata/project")` returns no problems, and the counts are 2 rules, 1 component, 1 template, 1 flow, and the expected token count.
- `TestLoadPointOfView`: missing design-system; design-system is a file; file missing; whitespace only (line 1); directory at the path.
- `TestParseTokens`: valid two-category file; number value keeps literal text (`4`, `1.5`); each error row in the table above with its line; invalid JSON reports the syntax error line; empty file; trailing content; duplicate category; duplicate token; name pattern failures (`Color`, `a_b`, `-a`).
- `TestParseTokensLineNumbers`: an error on the fourth line reports line 4.
- `TestParseRules`: zero entries; two valid entries with statement and reason; duplicate id; invalid ids (`Color`, `a..b`, `-x`, `a b`); no statement; no Reason; empty Reason; repeated Reason; fenced `## ` ignored; preamble text ignored; multi-line statement joined with a space.
- `TestLoadItems`: valid folder sorted by name; missing folder; folder is a file; non-`.md` and subdirectory ignored; bad kebab-case name reported and excluded; empty file; missing `# ` line; title is the heading text, trimmed.
- `TestScaffoldThenLoadIsValid`: `Scaffold` into `t.TempDir()`, then `Load` returns no problems and zero tokens and rules.
- `TestScaffoldRefusesExisting`: a pre-existing `design-system` file, and a pre-existing empty directory, each return `ErrAlreadyExists` and leave the original untouched.
- `TestProblemString`: with and without line.

**`cmd/tenet`**
- Keep the existing `parseRoot` tests.
- `TestRunDispatch`: `run([]string{"nope"}, ...)` returns 2 and mentions `unknown command`.
- `TestRunValidateFixture`: exit 0; stdout contains `design system valid`.
- `TestRunValidateProblems`: a temp root with a bad `tokens.json` on line 2 returns 1 and stdout contains `design-system/tokens.json:2:`.
- `TestRunValidateMissingDirectory`: exit 1, stdout contains `directory not found`.
- `TestRunValidateBadRoot`: `--dir` missing path, exit 2.
- `TestRunInit`: exit 0 in a temp root; then `runValidate` on that root exits 0; second `runInit` exits 1 and stderr mentions `already exists`.
- The serve path's startup summary is not unit-tested (it runs a stdio server). The README's piped-stdio check covers it, as in slice 1.

**`internal/server`**: unchanged. Existing tests use `testdata/project`, which is now a valid pack, so they still pass.

**Quality gates**, all of which must pass before the slice is committed:
- `go test ./...`
- `test -z "$(gofmt -l .)"`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`
- `go run ./cmd/tenet validate --dir testdata/project` exits 0
- Stdio check from the README still returns the point of view.

## Repository files

- `testdata/project/design-system/tokens.json`: `color` (`ink`, `paper`, `accent`) and `space` (`unit`). Literal values are the fixture's own, not Stripe's.
- `testdata/project/design-system/rules.md`: entries `color.tokens-only` and `spacing.unit-multiples`, each with a statement and a `Reason:` line.
- `testdata/project/design-system/components/button.md`, `templates/settings-page.md`, `flows/first-run.md`: one-paragraph items with `# ` titles.
- `README.md`: the format of each file (short, with the rules and an example), the three commands (`tenet`, `tenet validate`, `tenet init`), exit codes, and the note that `validate` is the check to run before connecting a client. The stdio check and the Claude Code connection stay.
- `CLAUDE.md`: unchanged here (`slice-build` B4 owns the Project section for slice 1 only; later slices propose additions at closing).

## Out of scope for slice 2

Any MCP tool that returns tokens, rules, components, templates, or flows (slice 3). Reference validation between items and rules or tokens. Rule-id references from critique (slice 4). Critique checklist file (ADR-0002 lists it; the roadmap does not assign it to a slice, so it stays out). Front matter, versioning, packs, the Stripe reference pack (slice 5). Size limits on tool responses (slice 3).

## Open risks

- The JSON line numbers are exact for one-line tokens. Multi-line nested values are reported at their opening line (D23); a reader may find that slightly off for a deep object.
- Unclosed code fences hide later rule headings silently (Error table). A user will see fewer rules than they wrote. Acceptable for slice 2; a later slice can add a check for an unclosed fence.
- The serve path logs problems but still serves. A user who never reads stderr will see an empty tool response for tokens later. Slice 3 must make the tools report pack problems.
