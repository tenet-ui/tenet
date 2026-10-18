# Consistency experiment report

## 1. Question

Does generating UI with the tenet tools and the reference pack make it more consistent than a plain prompt? This is the Slice 6 question on the roadmap, and it tests the risk named in ADR-0003.

## 2. Setup

- Date: 2026-10-18.
- Model: claude-sonnet-5-5.
- Claude CLI: 2.1.302 (Claude Code).
- Briefs: five (payments-list, empty-state, account-setup, sign-in, settings-card), one sample per brief per condition, so n=5 per condition.
- Pack: packs/stripe-reference.
- Plain condition: no MCP servers, no tools.
- Tools condition: tenet MCP server with `--dir packs/stripe-reference`, allowed tools `mcp__tenet__*`, plus the fixed addendum from briefs.md.
- Metrics (E12-E14):
  - Critique findings: findings produced by the pack's rules.
  - Token-only: zero findings AND at least one `var(--` reference.
  - Unresolved names: `var(--x` names that match no flattened token name.

## 3. Results

| Condition | Brief | Status | Token-only | Critique findings | Unresolved var names |
|---|---|---|---|---|---|
| plain | account-setup | ok | no | 41 | 19 |
| plain | empty-state | ok | no | 29 | 11 |
| plain | payments-list | ok | no | 52 | 36 |
| plain | settings-card | ok | no | 25 | 18 |
| plain | sign-in | ok | no | 37 | 22 |
| tools | account-setup | ok | no | 11 | 0 |
| tools | empty-state | ok | no | 9 | 0 |
| tools | payments-list | ok | no | 11 | 0 |
| tools | settings-card | ok | no | 11 | 0 |
| tools | sign-in | ok | no | 11 | 0 |

- plain: 0 token-only of 5; 184 findings in total; 36.8 findings per ok output; 5 ok, 0 failed
- tools: 0 token-only of 5; 53 findings in total; 10.6 findings per ok output; 5 ok, 0 failed

## 4. Failures

None. All ten runs produced an output; no .failed file exists.

## 5. Answer (pre-registered, E17)

Tools has 0 token-only outputs against plain's 0, so "more token-only" does not hold. Tools has 53 findings in total against plain's 184, so "fewer findings" holds. Exactly one holds, so the answer is MIXED.

## 6. Post-hoc finding (NOT part of the pre-registered answer)

This split was chosen after seeing the results. The metric was not changed, and the pre-registered answer in section 5 stands as written.

All 53 tools findings are color.tokens-only hits inside a `:root` block, where the output declares pack colour values as custom properties (for example `--color-ink: #0d1b2e;`) and then uses `var()`. No tools output has a raw colour outside a `:root` block.

Plain outputs have 116 findings inside their own `:root` blocks (the light block and the dark-mode block) and 68 outside them, at the use sites.

| Group | Findings in `:root` blocks | Findings at use sites (outside `:root`) |
|---|---|---|
| plain | 116 | 68 |
| tools | 53 | 0 |

The 68 plain use-site findings are 52 spacing, 13 radius and 3 colour. The use-site difference is therefore mostly spacing and radius values, not colour. The three plain colour use-site findings are the only colour findings outside a `:root` block in any output.

The token-only metric counts a token declaration as a raw value, so every tools output is marked not token-only even though its values come from the pack. On this post-hoc split, the delivery direction is supported at the use sites: the tools outputs use pack values there, and the plain outputs do not. The pre-registered token-only metric could not show that.

## 7. Limits

- Five briefs, one sample per condition: direction, not proof (E16).
- One model.
- Critique checks text against the pack's Forbid: patterns only. Judgement about the point of view and the user journey is not measured.
- Token-only is a proxy for raw values and named values, not for their correct use.
- Unresolved names depend on the flattened naming (E14).
- Run-setting changes from the plan:
  - `--bare` was not used, because it skips keychain reads and the subscription login then fails with "Not logged in" (smoke A). Without it, the user's global `~/.claude/CLAUDE.md`, auto-memory, and user plugins, skills and hooks load into both conditions. This is the very inclusion E4 meant to exclude. It applies equally to both conditions, so the bias is probably small, but it is not zero.
  - `--tools=""` is passed in equals form, because the space-separated form was rejected as "argument missing" by this CLI version.
- Runner behaviour, not exercised by any run in this experiment (no run failed):
  - The runner strips every line that starts with a code fence from a reply, not only one wrapping fence as E8 specifies. No output contains a fence, so this has no effect on the ten outputs.
  - On a non-zero exit, the `.failed` file records only stderr. Stdout is dropped, and `claude -p` sometimes writes its error there, so a failure can be recorded as the generic "claude exited non-zero".
- Brief wording: the settings-card brief reads "save, cancel, and delete the account". E11 asks for save as the primary action and delete as destructive. The brief is frozen after the first run, so the wording stays as run, and the report does not claim the destructive styling was tested.
