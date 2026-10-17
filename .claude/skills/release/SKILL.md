---
name: release
description: "Use when every slice above ## Later in docs/roadmap.md is ticked and the user wants to ship. Hardening, docs, version tag, deploy, smoke test."
argument-hint: "[version]"
model: sonnet
---

Release phase of the idea-to-software flow. There are no approval gates: do not stop for the human to verify or approve work. Stop and ask only when you cannot decide with strong confidence.

**Decide, don't ask.** Never ask the human to verify, confirm, or approve. Ask only when you cannot decide with strong confidence, and the decision is theirs alone (product intent, scope, money, accounts, secrets, outward-facing or hard-to-reverse actions); `CLAUDE.md` has the full rule. Decide everything else yourself, record it in the changelog or an ADR, and name the notable ones in your final summary.

This phase runs on Sonnet, whatever the session default. The `model:` override lasts only until the human's next message, so call the Skill tool with `release` again at the start of each turn that continues this phase.

## Steps

1. **Readiness.** Every slice above `## Later` in `docs/roadmap.md` is ticked. On `main`, run the full test suite and build, and recheck every slice's acceptance criteria in the running app with `run`.
2. **Security.** Call the Skill tool with `security-review` over the whole codebase; process findings with `superpowers:receiving-code-review`.
3. **Hardening.** Config and secrets come from the environment, and `git log -p` holds no secrets. The core workflow has logging and handles its error paths.
4. **Docs.** README covers what it is, install, configure, run, and test. `CHANGELOG.md` lists this release.
5. **Version.** Propose the version (`$ARGUMENTS` or semver from the changelog) then bump, commit, and tag.
6. **Ship.** Deploy or package per the deploy ADR. For steps only the human can do, call the Skill tool with `wizard`.
7. **Smoke test** the released artifact itself (deployed URL, installed package, built binary): walk the core workflow end to end.
8. **Next.** Move `## Later` items into a new roadmap via the `foundation` skill, step A3.

**Done when:** the tagged release passes its smoke test and you have confirmed it yourself.
