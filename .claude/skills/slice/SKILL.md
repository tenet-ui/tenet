---
name: slice
description: "Use when starting or resuming the design of a slice from docs/roadmap.md, before its spec is approved. Orients on the slice and designs it into an approved spec."
argument-hint: "[slice number]"
model: opus
---

Design half of the build loop, run once per **vertical slice** in `docs/roadmap.md`. The build half is the `slice-build` skill. There are no approval gates: do not stop for the human to verify or approve work. Stop and ask only when you cannot decide with strong confidence.

**Decide, don't ask.** Never ask the human to verify, confirm, or approve. Ask only when you cannot decide with strong confidence, and the decision is theirs alone (product intent, scope, money, accounts, secrets, outward-facing or hard-to-reverse actions); `CLAUDE.md` has the full rule. Decide everything else yourself, record it in the spec, and name the notable ones in your final summary.

This phase runs on Opus, whatever the session default. The `model:` override lasts only until the human's next message, so call the Skill tool with `slice` again at the start of each turn that continues this phase.

## B0. Orient

1. Read `GLOSSARY.md`, `docs/adr/`, and `docs/roadmap.md`. The slice is `$ARGUMENTS`, or the first unticked one.
2. When the slice's spec in `docs/superpowers/specs/*slice-<N>-*` already carries an `Approved:` line, call the Skill tool with `slice-build` and stop here.
3. Use the slice's acceptance criteria from the roadmap as the definition of done for the design.

## B1. Design

Call the Skill tool with `superpowers:brainstorming` for this slice, topic `slice-<N>-<name>`. Feed it the context from B0 and tell it to answer its own clarifying questions and self-approve each design section, asking the human only where it cannot decide with strong confidence. The spec covers every acceptance criterion plus error and edge cases, and names each module and interface the slice adds, so the plan can be written without design choices left open.

**Done when:** the spec is complete and self-reviewed against the acceptance criteria. Then:

1. Add `Approved: <YYYY-MM-DD>` under the spec's title and commit.
2. Call the Skill tool with `slice-build` in this same turn (it switches to Sonnet). It takes over from brainstorming's hand-off to writing-plans.
