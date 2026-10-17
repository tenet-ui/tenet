---
name: foundation
description: "Use when the user shares a new product idea, or the repo lacks GLOSSARY.md or docs/roadmap.md. Grills the idea into scope, glossary, ADRs, and a roadmap of vertical slices."
argument-hint: "[idea]"
model: opus
---

Foundation phase of the idea-to-software flow. Every step ends in a committed file; those files carry the project across sessions. There are no approval gates: do not stop for the human to verify or approve work. Stop and ask only when you cannot decide with strong confidence.

**Decide, don't ask.** Never ask the human to verify, confirm, or approve. Ask only when you cannot decide with strong confidence, and the decision is theirs alone (product intent, scope, money, accounts, secrets, outward-facing or hard-to-reverse actions); `CLAUDE.md` has the full rule. Decide everything else yourself, record it in the glossary, an ADR, or the roadmap, and name the notable ones in your final summary.

Resume at the first step whose output is missing. This phase runs on Opus, whatever the session default. The `model:` override lasts only until the human's next message, so call the Skill tool with `foundation` again at the start of each turn that continues this phase.

## A0. Skills setup

If the directory is not a git repository, run `git init -b main` and make an initial commit; every later step commits and branches. If `docs/agents/` is missing, run `setup-matt-pocock-skills` in place (see Running things in `CLAUDE.md`). It records the issue tracker and the `GLOSSARY.md`/ADR layout that `code-review` and `domain-modeling` read. Its recommended answers: GitHub when a GitHub remote exists, local markdown otherwise.

**Done when:** `docs/agents/` exists and is committed.

## A1. Grill with docs

First, size the idea. When it is too big or foggy to settle in one session (a large greenfield system, many unknown subsystems), run `wayfinder` in place instead; once its map clears, resume here at A3 with the map's decisions as input.

Otherwise, call the Skill tool twice, for `grilling` and `domain-modeling`, on the idea (from `$ARGUMENTS`, or ask for it). This is the `/grill-with-docs` pairing: the interview runs while `GLOSSARY.md` and `docs/adr/` are written the moment each term or decision crystallises. The design tree must reach every branch below:

- the problem, and the primary user who has it
- the single core workflow, start to finish
- **MVP scope**, plus an explicit **non-goals** list
- success criteria, each observable or measurable
- constraints: platform, stack, deploy target, data and persistence, auth, external APIs, budget
- the biggest unknowns and risks

When a question needs someone else's knowledge, run `to-questionnaire` in place and park that branch until the answers return.

**Done when:** every branch above has an answer (questions go to the human only where you cannot decide with strong confidence), and every expensive-to-reverse decision (stack, architecture style, storage, auth, external dependencies) has an ADR. Commit.

## A2. Spike the risks (only for risks named in A1)

For each risk that could invalidate the roadmap, call the Skill tool with `prototype` to answer that one question. Commit the prototype to a `prototype/<name>` branch out of main, kept as a primary source. Record the verdict as an ADR that points at that branch.

**Done when:** every listed risk has an ADR with its verdict.

## A3. Roadmap

Write `docs/roadmap.md` from `GLOSSARY.md` and the ADRs, in exactly this shape (`CLAUDE.md` and the slice skills parse it):

```markdown
# Roadmap

- [ ] Slice 1: Walking skeleton
  - <acceptance criterion a human can check by hand>
- [ ] Slice 2: <user-visible capability>
  - <acceptance criterion>

## Later
- <out-of-MVP idea>
```

- **Slice 1 is the walking skeleton**: scaffold with a `.gitignore` for the chosen stack, one thin end-to-end path through the core workflow, test runner, linter/formatter, CI, README with run instructions.
- Every later slice is a **vertical slice**: one capability a user can see and demo, cutting through every layer it needs.
- Non-goals and nice-to-haves go under `## Later`.

**Done when:** the roadmap is committed.

## Hand-off

Note the foundation is done in your summary, then call the Skill tool with `slice` to start designing Slice 1.
