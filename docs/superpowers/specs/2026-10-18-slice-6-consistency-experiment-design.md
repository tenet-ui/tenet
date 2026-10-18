# Slice 6: Consistency experiment — design

Approved: 2026-10-18

Self-approved under the project's no-approval-gates rule (see CLAUDE.md, "Decide, don't ask"). Path: architectural (a new experiment subsystem with its own folder, runner and tally program; no product code changes).

## Goal

Measure whether generated UI built with tenet's delivery tools and the reference pack is more consistent than UI generated from a plain prompt. Consistency is measured two ways: whether an output uses only pack tokens, and how many `critique` findings it produces against the pack's rules. The result is a committed report that answers the question and settles the risk the roadmap attaches to ADR-0003.

## Acceptance criteria (from docs/roadmap.md)

1. Five fixed briefs are run twice each: once with a plain prompt, once with the MCP tools enabled and the reference pack loaded.
2. A committed report lists how many generated outputs used only pack tokens and how many critique findings each produced, and states whether tool delivery improved consistency (this settles the risk in ADR-0003).

Read "run twice each" as one plain run and one tools run per brief, giving 10 runs. Each brief is run in both conditions; the roadmap does not ask for repeated samples per condition.

## Decisions

| # | Decision | Reason |
|---|---|---|
| E1 | The experiment lives in `experiments/consistency/`. It adds no code under `internal/` or `cmd/`. | The product does not change for an experiment. The experiment's code and data stay separate and reviewable. |
| E2 | The folder holds: `briefs.md` (the five fixed briefs and the two prompt templates), `run.sh` (the runner), `outputs/plain/*.html` and `outputs/tools/*.html` (raw generated pages), `tally/` (a Go program), and `REPORT.md`. | One folder, one place to rerun or audit. |
| E3 | Generation model: `claude-sonnet-5-5`, pinned with `--model` for all 10 runs. | One fixed model makes the runs comparable. Sonnet is the project's default model for implementation work and keeps the cost of 10 page generations low. |
| E4 | Each run starts from a `mktemp -d` directory outside the repository tree, holding only the MCP config and the built binary. `claude` walks up parent directories for `CLAUDE.md`, so a directory inside the repo would load the project's file; a directory outside it does not. The runner also passes `--bare` if the CLI supports it (checked in the first run, E22), so the user's global `CLAUDE.md` and auto-memory are not loaded either. The runner removes the directory at the end. | Without this, the repo's `CLAUDE.md` loads into the plain run and biases it toward the project's rules. |
| E5 | Plain condition: `claude -p --model claude-sonnet-5-5 --strict-mcp-config --mcp-config '{"mcpServers":{}}' --tools "" --no-session-persistence --max-budget-usd 2`, with the plain prompt on stdin. | No MCP servers and no tools, so the only input is the brief. |
| E6 | Tools condition: the same flags, with `--mcp-config <file>` pointing to a JSON file that starts the built binary `<abs>/tenet --dir <abs>/packs/stripe-reference` as server `tenet`, and `--allowedTools "mcp__tenet__*"`. The tenet binary is built by the runner with `go build -o <work>/tenet ./cmd/tenet`. | The condition under test is "MCP tools enabled and the reference pack loaded". Pointing `--dir` at the pack loads exactly that system. |
| E7 | Both conditions use the same brief text. The tools condition adds one fixed paragraph (in `briefs.md`) telling the model that the project's design system is available through the `tenet` tools, that it must fetch the point of view and the components, template or flow that fit the brief before writing the page, and that it must build only from those tokens. The plain condition gets no such paragraph. | The only difference between conditions is the tool access and the instruction to use it. |
| E8 | Output contract, shared by both conditions: reply with one complete HTML document, with CSS inline in a `<style>` element, no external assets, no scripts that load anything, and no commentary. The runner saves the reply to `outputs/<condition>/<brief-id>.html` after removing one wrapping Markdown code fence if present. | Critique reads text files, and a single file per output keeps the tally simple. |
| E9 | Retry and failure: a run that returns an error, or whose saved output contains no `<html` element, is retried once. If the retry also fails, the runner writes `outputs/<condition>/<brief-id>.failed` with the error text and no HTML file. A failed output counts as not token-only and has no critique count; the report lists it. | Failures must be visible, not dropped. |
| E10 | Budget: `--max-budget-usd 2` on every run. A run that hits the cap is a failure under E9. | Caps the cost of a single runaway run. |
| E11 | Briefs (five, project-neutral, each mapped to pack items): `payments-list` (a page listing payments with amount, customer, status and date, with a filter and an export action; maps to template `payments-list`); `empty-state` (an empty state for a payments list that has no payments yet; maps to component `empty-state`); `account-setup` (the first step of account setup, asking for business name and country, with a continue and a back action; maps to flow `account-setup`); `sign-in` (an email and password sign-in form with a primary sign-in button and a secondary create-account action; maps to components `text-field` and `button`); `settings-card` (a card with a title and a description and three actions: save as primary, cancel, and a delete action that is destructive; maps to component `button`). | Each brief exercises a pack item, so the pack is actually relevant to every brief. The briefs do not name the pack or the tenet tools, except in the tools-only paragraph. |
| E12 | Metric 1, critique findings: the tally program loads the pack with `designsystem.Load("packs/stripe-reference")`, reads each saved output, and calls `designsystem.Critique(source, pack.Rules)`. The count is `len(findings)`. Only rules that carry `Forbid:` lines produce findings (slice 4 C4). | Reuses the product's own checker, so the experiment measures the same thing a user would see from `critique`. |
| E13 | Metric 2, token-only: an output is token-only when it has zero critique findings AND contains at least one `var(--` reference. A failed run is not token-only. | Zero findings alone can be vacuous (a page with no styling has no raw values). The `var(--` check requires that the page actually uses named values. |
| E14 | Metric 3, unresolved var names (reported, not gating): the tally flattens `tokens.json` into CSS-style names by joining the path with `-` and prefixing `--` (for example `color.ink` becomes `--color-ink`, `space.md` becomes `--space-md`, `motion.standard` becomes `--motion-standard`). Each `var(--name)` in an output whose name is not in this set is counted as unresolved. The count is shown per output. | Pack tokens have no CSS naming convention, so models may invent names. The report shows this as its own finding about delivery rather than hiding it inside the token-only test. |
| E15 | The tally emits a Markdown table with one row per output: condition, brief, status (ok or failed), token-only (yes or no), critique findings, and unresolved var names. It also prints a per-condition summary: count of token-only outputs, total findings, and mean findings per output. The tally does not write `REPORT.md`; the report is written from the tally output. | The numbers come from one reproducible program. The prose is written once, from those numbers. |
| E16 | Sample size: one output per brief per condition, so n=5 per condition. The report states that five briefs and one sample each cannot establish a statistical effect, and it answers the roadmap's question with the observed numbers and that caveat. | The roadmap asks for a direction-setting answer; the report must not overstate it. |
| E17 | Result answer: the report states "improved" only if the tools condition has more token-only outputs AND fewer total findings than the plain condition. Otherwise it states "did not improve" or "mixed", with the numbers. | A rule fixed before the run, so the answer is not chosen after seeing it. |
| E18 | ADR-0003 gets a dated "Result" section at its end, linking `experiments/consistency/REPORT.md` and stating the outcome in one or two sentences. The ADR's decision is not changed unless the report contradicts it; in that case a new ADR supersedes it, per the slice flow. | The roadmap says the experiment settles the ADR-0003 risk. The record must point at the evidence. |
| E19 | Roadmap: Slice 6 is ticked (`- [x]`) once the report is committed and `go test ./...` passes. | Matches how slices 1 to 5 were recorded. |
| E20 | The tally program lives in `experiments/consistency/tally/` as `package main`, imports `github.com/tenet-ui/tenet/internal/designsystem` (allowed: the import path is under the module root), and has one unit test for the token check (E14's flattening and unresolved-name counting). It must pass gofmt and golangci-lint. | The one piece of logic in the experiment is tested. The lint rule applies to all Go in the module, so the tally follows the same standard. |
| E21 | The generated outputs (`outputs/`) are committed. The tenet binary and the MCP config live only in the temp directory from E4 and are not committed. | The outputs are the evidence; the binary is a build artifact. |
| E22 | Cost and spend: the runner's total budget is up to 10 runs at no more than 2 USD each (20 USD maximum). The first run is checked by hand before the rest are started. | The spend is bounded and visible before it is committed to. |

## Architecture

```
experiments/consistency/
├── briefs.md               # five briefs; plain and tools prompt templates (E7, E8, E11)
├── run.sh                  # builds tenet, writes the MCP config, runs 10 claude -p calls (E3–E10)
├── outputs/
│   ├── plain/<brief-id>.html|.failed
│   └── tools/<brief-id>.html|.failed
├── tally/
│   ├── main.go             # loads the pack, runs Critique, prints the tables (E12–E15)
│   └── main_test.go        # token-name flattening and unresolved-name counting (E14, E20)
└── REPORT.md               # written from the tally output (E15–E17)
docs/adr/0003-on-demand-delivery-through-tools.md   # Result section (E18)
docs/roadmap.md                                       # slice 6 ticked (E19)
```

The runner's temp directory (E4) is outside the repository and is removed at the end of the run.

Dependency direction is unchanged. `experiments/consistency/tally` depends on `internal/designsystem`. Nothing in `internal/` or `cmd/` depends on `experiments/`.

The runner is a shell script because it only calls the CLI, builds one binary and writes files. The tally is Go because it reuses `designsystem` and must pass the same lint and format checks as the product.

## Interfaces

**`experiments/consistency/tally`**

```go
// flattenTokens returns the CSS-style names of the tokens in tokens.json,
// for example "--color-ink" for color.ink. Numbers and strings are both named.
func flattenTokens(data []byte) (map[string]bool, error)

// unresolvedNames returns each var(--name) name in source that is not in names, in order of first appearance, duplicates included.
func unresolvedNames(source string, names map[string]bool) []string

// hasVarRef reports whether source contains at least one var(-- reference.
func hasVarRef(source string) bool
```

`main` runs the pipeline: load the pack, flatten its tokens, and for each output in both condition folders, run `Critique`, count unresolved names, and decide token-only. Failed outputs (a `.failed` file) are listed as failed without a count.

**`experiments/consistency/run.sh`** takes no arguments. It writes its progress to stderr and exits non-zero if any run fails after its retry, after finishing the others.

## Data flow

1. `run.sh` builds `<work>/tenet` from the repo, writes `<work>/mcp.json`, and extracts each brief's prompt from `briefs.md`.
2. For each condition and brief, it runs `claude -p` (E5 or E6) from `<work>`, with the prompt on stdin, and saves the reply under `outputs/<condition>/` (E8). A failed run is retried once (E9).
3. `go run ./experiments/consistency/tally` reads the pack and the saved outputs, and prints the per-output table and per-condition summary (E15).
4. `REPORT.md` is written from that output, with the answer from E17 and the caveat from E16.
5. ADR-0003 and the roadmap are updated, and the slice is committed.

## Errors

| Situation | Result |
|---|---|
| `claude` not on PATH | `run.sh` exits with `claude CLI not found` before any run. |
| `go build` of the tenet binary fails | `run.sh` exits non-zero before any run. |
| A run returns an error, or its output has no `<html` | Retried once (E9). If it fails again: a `.failed` file with the error text. |
| A run hits the budget cap | Treated as a run error (E9, E10). |
| Pack has load problems | The tally prints the problems and exits non-zero; no table is printed (a broken pack would make the critique counts meaningless). |
| `tokens.json` does not parse in the tally | The tally exits non-zero with the parse error. |
| An output file cannot be read | The tally exits non-zero naming the file. |

## Testing

- `tally/main_test.go`, table-driven:
  - `flattenTokens` on a small JSON with nested objects, a number and a string returns the expected names, including `--space-md` and `--motion-standard`.
  - `unresolvedNames` returns names missing from the set, in order, with duplicates, and returns none when all names resolve.
  - `hasVarRef` is true for `var(--color-ink)` and false for `color: #fff`.
- The critique counts are not unit-tested here: they call `designsystem.Critique`, which slice 4 already tests.
- Full suite: `go test ./...`, `test -z "$(gofmt -l .)"`, and golangci-lint (the commands in CLAUDE.md), run after the tally is added.
- The experiment itself is verified by reading the outputs: the first plain and first tools run are checked by hand (E22) before the other eight are started, and the failed-run path is exercised once by running the runner with a bad model name in a test run whose output is discarded.

## Known limits (stated in REPORT.md)

- Five briefs, one sample per condition (E16). The result is direction, not proof.
- One model (E3). Results for other models are not tested.
- Critique checks text against the rules' `Forbid:` patterns only (slice 4 C6, C14). Judgement about the point of view or the user journey is not measured.
- Token-only (E13) is a proxy: it checks for raw values and for the presence of named values, not that each named value is the right token for its role.
- Unresolved names (E14) depend on the naming convention chosen for the flattened tokens. The report shows the count, not a verdict on naming.
