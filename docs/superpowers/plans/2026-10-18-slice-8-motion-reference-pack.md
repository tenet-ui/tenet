# Slice 8: Motion reference pack Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Do not commit; leave changes in the working tree.

**Goal:** Ship `packs/motion-reference/` and a clean Remotion fixture, with tests, so `critique` checks motion source against an invented motion personality.

**Architecture:** Content only. The pack uses the existing format (ADR-0011); no change to `internal/designsystem` or `internal/server`. Tests are a new file in `internal/designsystem` that follows `pack_test.go`.

**Tech Stack:** Go, Markdown, JSON, one TSX fixture (not compiled).

**Spec:** `docs/superpowers/specs/2026-10-18-slice-8-motion-reference-pack-design.md` (decisions M1 to M12). Read it and `packs/stripe-reference/` first; copy its file shapes.

## Global Constraints

- Pack root `packs/motion-reference/`, design system at `packs/motion-reference/design-system/`.
- Tokens exactly as spec M3; token names must not match the pack's own patterns (M4).
- Nine rules: five checkable (`motion.no-builtin-easing`, `motion.springs-settle`, `motion.time-from-tokens`, `motion.stagger-from-tokens`, `color.tokens-only`), four judgement (`motion.one-focus`, `motion.reading-order`, `motion.hold-the-payoff`, `motion.exit-shorter-than-enter`). Patterns exactly as spec M6; every rule has `Reason:` and `Applies to:`; checkable rules have `Fix:`.
- Three components (`title-reveal`, `number-counter`, `chart-draw-on`), template `feature-showcase`, flow `launch-storyboard`. Every item file ends with a `Source:` line saying the personality is invented.
- Only text files in the pack. No existing file outside the new ones is modified.

## Review Focus

- `ease` as a whole word only: `release` and `increase` in source must not match (test in Task 2).
- `fps={fps}` and `durationInFrames={frames}` (token expressions) must not match `motion.time-from-tokens` (Task 2).
- `Easing.bezier(...settle)` must not match; `Easing.bezier(0.16, 1, 0.3, 1)` must (Task 2).
- A CSS custom property `--motion-settle: cubic-bezier(0.16, 1, 0.3, 1);` must not be flagged (Task 2).
- No pack file other than `rules.md` triggers the pack's own patterns (Task 2).

---

### Task 1: Pack content and load test

Model: sonnet

**Files:**
- Create: everything under `packs/motion-reference/` (see spec Architecture)
- Create: `internal/designsystem/motion_pack_test.go`

**Interfaces:**
- Produces: consts `motionRoot = "../../packs/motion-reference"` and `motionFixture = "../../testdata/critique/motion-clean.tsx"`, and `loadMotionPack(t *testing.T) DesignSystem` (fails the test on any problem), used by Task 2.

- [ ] **Step 1: Write failing test** `TestMotionPackLoadsClean`: `loadMotionPack`; assert 3 components, 1 template, 1 flow, 9 rules, `Checkable(pack.Rules) == 5`. Also `TestMotionPackShipsOnlyText` (extensions `.md`, `.json` only) and `TestMotionReadmeStatesInvented` (lowercased README contains `invented`, `no outside source`, `remotion`, `css`).
- [ ] **Step 2:** `go test ./internal/designsystem -run TestMotion` fails (pack missing).
- [ ] **Step 3:** Write the pack files per spec M2, M3, M5 to M9. Point of view in the voice of Settled; README with the invented label, token table, the `tokens.ts` convention (M7), how to validate and critique. Rule `Reason:` lines explain the personality, not a research section.
- [ ] **Step 4:** Run `go test ./internal/designsystem -run TestMotion` and `go run ./cmd/tenet validate --dir packs/motion-reference`; both pass.

### Task 2: Fixture and rule-behaviour tests

Model: sonnet

**Files:**
- Create: `testdata/critique/motion-clean.tsx`
- Modify: `internal/designsystem/motion_pack_test.go`

**Interfaces:**
- Consumes: Task 1's `loadMotionPack`, `motionFixture`.

- [ ] **Step 1: Write failing tests:** `TestMotionFixtureHasNoFindings` (`Critique(fixture, pack.Rules)` is empty); `TestMotionSeededViolations` table, each edit appended to the fixture text and expected to give exactly one finding with the named rule ID: `transition: opacity 300ms var(--motion-settle);` gives `motion.time-from-tokens`; `transition: opacity var(--motion-duration-base) linear;` gives `motion.no-builtin-easing`; `Easing.bezier(0.16, 1, 0.3, 1)` gives `motion.no-builtin-easing`; `spring({ config: { damping: 8 } })` gives `motion.springs-settle`; `<Sequence from={0} durationInFrames={90}>` gives `motion.time-from-tokens`; `const delay = 12;` gives `motion.stagger-from-tokens`; `color: #ff0000;` gives `color.tokens-only`. `TestMotionPatternsOnSnippets` table covering the Review Focus lines above (no findings for `release`, `increase`, `fps={fps}`, `Easing.bezier(...settle)`, the custom-property declaration). `TestMotionPackDoesNotTriggerItself`: run `Critique` over every pack file except `design-system/rules.md` and expect no findings.
- [ ] **Step 2:** Run; fails (fixture missing).
- [ ] **Step 3:** Write `motion-clean.tsx`: a Remotion composition (title-reveal plus counter using `spring`, `interpolate`, `Easing.bezier(...settle)`, `Sequence`) importing from `./tokens`; uses only token expressions. If a seeded test shows a pattern wrong, fix the pattern in `rules.md`, not the test, unless the test contradicts the spec.
- [ ] **Step 4:** `go test ./...`, `test -z "$(gofmt -l .)"`, and the lint command from CLAUDE.md pass.
