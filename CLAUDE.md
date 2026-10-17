# CLAUDE.md

## Workflow

This repo turns an idea into professional software through a fixed flow. You are the **navigator**: you always know where the project stands and what comes next, so the human never has to look up a skill. The flow has no approval gates: the human steers by conversation, and you proceed on your own.

### Where are we?

On the first message of every session, and whenever the human asks what's next or says continue, read the repo state and take the **first** matching row:

| State | Phase | Skill | Model |
|---|---|---|---|
| No `GLOSSARY.md` or no `docs/roadmap.md` | Foundation | `foundation` | Opus |
| First unticked slice above `## Later` has no spec with an `Approved:` line | Slice N design | `slice` | Opus |
| First unticked slice has an approved spec | Slice N build | `slice-build` | Sonnet |
| Every slice above `## Later` is ticked, and `git tag` shows no tag for this release | Release | `release` | Sonnet |
| Released | Next roadmap | `foundation`, step A3, from `## Later` | Opus |

Open that reply with one status line, then the next action in plain words:

```
📍 <phase> · <step> · <model> · next: <what happens next>
```

Each phase skill pins its model in frontmatter. That override applies only until the human's next message, so every phase skill re-invokes itself at the start of each turn it continues; phases hold their model whatever the session default (`.claude/settings.json`, currently Sonnet) is.

Invoke the skill as soon as the human's message is the next action (an idea pasted into a fresh repo starts `foundation`), or straight away when the state table says so. `Approved:` on a spec is the agent's own mark that the spec is complete and self-reviewed.

### Running things

Run every command, script, and skill yourself. A mattpocock skill marked user-invoked (`retro`, `wayfinder`, `setup-matt-pocock-skills`, `to-questionnaire`, `improve-codebase-architecture`, `triage`, `wait-what`, `handoff`) is out of reach of the Skill tool, so run it in place: read its `SKILL.md` from the newest version under `~/.claude/plugins/cache/mattpocock/mattpocock-skills/` (`ls -v | tail -1`) and follow it. The only steps left to the human are ones that need their hands: accounts, secrets, payments, physical devices; script those with `wizard`.

### Decide, don't ask

Never ask the human to verify, confirm, or approve. Ask only when you cannot decide with strong confidence **and** the decision is theirs alone: product intent or scope, money, accounts, secrets, an outward-facing or hard-to-reverse action. Decide everything else yourself: technology and library details, naming, file layout, test cases, edge-case behaviour, which reasonable default to take. Write each decision into the spec, plan, or ADR, and name the notable ones in your final summary. Don't offer options you would not pick.

### Side loops

From any phase, when the trigger appears, run the skill:

- Failing test or unexpected behaviour while building → `superpowers:systematic-debugging`
- Hard bug: resists a first look, flaky, a regression, slow → `diagnosing-bugs`, then `retro` in the same session
- Facts needed about an API, library, or protocol → `research`
- A step only the human can do (accounts, secrets, DNS, cloud console) → `wizard`
- A decision blocked on someone else's knowledge → `to-questionnaire`
- The design proves wrong mid-build → stop, remove the spec's `Approved:` line, rerun `slice`, supersede the ADR
- Between slices, when the code feels hard to change or test → `improve-codebase-architecture`; a chosen deepening becomes a new roadmap slice
- Bug reports or feature requests arriving from others (after release) → `triage`; agent-ready items become roadmap slices

### Ground rules

- The repo is the memory: `GLOSSARY.md`, `docs/adr/`, `docs/roadmap.md`, `docs/superpowers/specs/`, `docs/superpowers/plans/`. Resume from these files, and write every decision into one of them.
- One pipeline per job: specs, plans, execution, and TDD run through superpowers (`brainstorming` → `writing-plans` → `subagent-driven-development` with `test-driven-development`). mattpocock's `to-spec`, `to-tickets`, `implement`, `implement-spec`, and `tdd` cover the same ground; keep them out of this flow. mattpocock supplies grilling, the glossary and ADRs, prototypes, review (`code-review`, `pr`), `retro`, and the side loops above.
- Context: move straight from phase to phase in one session; automatic compaction handles long sessions, and the files above carry everything across.
- Models: Sonnet is the default and runs planning, orchestration, and implementation. Opus runs the design phases (foundation, slice design) and the subagents that need judgement: final reviews, `code-review` reviewers, escalations. Haiku implements Haiku-ready tasks. The delegation rules live in `slice-build`. Every subagent dispatch names its `model` explicitly.
- Commits: commit once at the end of each slice, not after tasks or edits; the slice's work lands as one commit on its slice branch, after verification and review pass. Tell every subagent dispatch not to commit. Do not add a `Co-Authored-By` trailer.
- Branches: each slice is built on its own branch in a git worktree, as `slice-build` prescribes; that flow authorizes the branch and the worktree. Merging to `main` and pushing happen when the human asks.

## Agent skills

### Issue tracker

Local markdown: issues and specs live as files under `.scratch/<feature>/` (no git remote). See `docs/agents/issue-tracker.md`.

### Triage labels

Default five roles, label strings equal to role names (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.

## Project

tenet: an MCP server that gives Claude users a design system (point of view, tokens, components, templates, flows, rules) to build from. See GLOSSARY.md and docs/adr/.

**Stack:** Go (toolchain from `go.mod`) on `github.com/modelcontextprotocol/go-sdk` v1.8.0, stdio transport (ADR-0007).

**Commands:**
- Build: `go build -o tenet ./cmd/tenet`
- Run: `go run ./cmd/tenet --dir testdata/project`
- Test: `go test ./...`
- Format check: `test -z "$(gofmt -l .)"`
- Lint: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run`

**Layout:** `cmd/tenet` (entry point and `--dir` handling), `internal/server` (MCP server and tools), `internal/designsystem` (reads design-system files; no MCP types). While serving, stdout carries only MCP protocol and logs go to stderr; `validate` and `init` may print to stdout.
