---
name: slice-build
description: "Use when a slice's spec carries an Approved: line and its plan, execution, verification, review, or merge is pending. Plans, delegates, verifies, reviews, and merges one slice."
argument-hint: "[slice number]"
model: sonnet
---

Build half of the build loop for one **vertical slice**, after its spec is approved (the `slice` skill designs it). There are no approval gates: do not stop for the human to verify or approve work. Stop and ask only when you cannot decide with strong confidence.

**Decide, don't ask.** Never ask the human to verify, confirm, or approve. Ask only when you cannot decide with strong confidence, and the decision is theirs alone (product intent, scope, money, accounts, secrets, outward-facing or hard-to-reverse actions); `CLAUDE.md` has the full rule. Decide everything else yourself, record it in the plan or an ADR, and name the notable ones in your final summary.

This phase runs on Sonnet, whatever the session default. The `model:` override lasts only until the human's next message, so call the Skill tool with `slice-build` again at the start of each turn that continues this phase.

## Orient

Read the slice's approved spec, `GLOSSARY.md`, and `docs/adr/`. Resume at the first step whose output is missing: plan in `docs/superpowers/plans/*slice-<N>-*`, unchecked plan tasks, a slice branch not yet merged.

## Delegation

Implementation is always delegated to subagents; this session plans, dispatches, and adjudicates. Every dispatch names its `model` explicitly:

| Work | Model |
|---|---|
| Implementer, task labelled `Model: haiku` | `haiku` |
| Implementer, task labelled `Model: sonnet` | `sonnet` |
| Per-task spec and quality reviews, scoped re-reviews | `sonnet` |
| Final slice review, and both `code-review` sub-agents | `opus` |
| Escalation: a task BLOCKED, or failing its fix loop | one tier up: haiku → sonnet → opus |

A task is **Haiku-ready** when every one of these holds; otherwise it is `sonnet`:

- it touches at most two files, each named by exact path
- its inputs, outputs, and error cases are spelled out (signatures, types, example values)
- its test cases are listed, with the exact command that runs them
- no design choice is left to the implementer: no new module, interface, or dependency to invent

**Commits.** Every implementer, fixer, and re-dispatch brief says: do not commit; leave changes in the working tree. Plan tasks carry no commit step. Once verification and review pass, make one commit for the whole slice on the slice branch (step 3 of the closing steps). `CLAUDE.md` holds the rule. With no commits, `review-package` sees an empty range: build each review diff with `git add -N .`, then `git diff -U10 --output=<file> <base>`.

## B2. Plan

Call the Skill tool with `superpowers:writing-plans`, plan named `slice-<N>-<name>`. When the slice introduces new modules or boundaries, call the Skill tool with `codebase-design` and apply its deep-module vocabulary to the plan before presenting it. Each task names exact files, starts with a failing test, ends with a verification command, and carries a `Model: haiku` or `Model: sonnet` line judged against the Haiku-ready test above. Fold scaffolding and config work (module init, ignore files, small loaders) into the task whose deliverable needs it: a task is worth its own implementer and reviewer only when a reviewer could meaningfully reject it while approving its neighbour.

**Done when:** the plan is written and self-reviewed, including each task's model label. Execution is subagent-driven, so implementation stays delegated.

## B3. Execute

Call the Skill tool with `superpowers:subagent-driven-development`. Its Model Selection follows the Delegation table above: the task's `Model:` line decides the implementer. It sets up the worktree, drives each task through `superpowers:test-driven-development` with spec and quality review. Its own final whole-branch review is B6: skip that dispatch, and stop at the last task's review. Routine test failures inside a task go through `superpowers:systematic-debugging`; a hard bug (resists a first look, flaky, a regression, slow) goes to the `diagnosing-bugs` skill.

A task still BLOCKED after escalating to opus goes back to the human with the plan change it needs.

**Done when:** every plan task is checked and each task's review is clean.

## B4. CLAUDE.md

- Slice 1: replace the placeholder under `## Project` in `CLAUDE.md` with the walking skeleton's real stack, build, test, lint, and run commands. Leave `## Workflow` untouched.
- Later slices: propose additions for conventions that settled during this slice; apply them directly.

## B5. Verify

Run these **before** `superpowers:finishing-a-development-branch` presents its options, even when execution hands off to it directly:

1. `superpowers:verification-before-completion`: full test suite, linter, and build, with output read.
2. `run`: launch the app and walk through every acceptance criterion of the slice.
3. CI is green on the branch (once CI exists).

**Done when:** each acceptance criterion has been observed working in the running app.

## B6. Review

1. Call the Skill tool with `code-review` (mattpocock-skills), fixed point `main`, spec path = this slice's spec in `docs/superpowers/specs/`. It reviews two axes in parallel (Standards and Spec); dispatch both sub-agents with model `opus`.
2. Call the Skill tool with `security-review` when the slice touches auth, user data, payments, file uploads, or network input.
3. Process findings with `superpowers:receiving-code-review`: verify each one, fix the real ones, push back on the wrong ones with evidence.
4. Rerun B5 after any fix.

**Done when:** every finding is fixed or rejected with a stated reason, and B5 passes again.

## B7. Finish

Call the Skill tool with `superpowers:finishing-a-development-branch` (merge, PR, keep, or discard). The PR route needs a GitHub remote; write the PR body with the `pr` skill.

## B8. Close

1. Tick the slice in `docs/roadmap.md`; add discoveries under `## Later`.
2. Add new terms to `GLOSSARY.md` and an ADR for each new decision.
3. Commit, then run `retro` in place on this session (see Running things in `CLAUDE.md`); it proposes environment fixes: checks, standards, navigation pointers. Apply the ones you judge worthwhile.
4. When `retro` finds a bug had no good seam to lock it down, run `improve-codebase-architecture` in place before the next slice.
5. Continue with the next slice: call the Skill tool with `slice`, or `release` when every slice above `## Later` is ticked.
