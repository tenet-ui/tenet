# Slice 8: Motion reference pack — design

Approved: 2026-10-18

Self-approved under the project's no-approval-gates rule (see CLAUDE.md, "Decide, don't ask"). Path: architectural in size (a new pack, two fixtures, one test file), but no change to `internal/designsystem` or `internal/server` code, no new tool and no new exported type (ADR-0011).

## Goal

Ship `packs/motion-reference/`, a worked motion pack: an invented motion personality with tokens, rules, components, one scene template and one storyboard flow, plus a Remotion fixture that `critique` finds clean. It shows that the existing pack format, delivery tools and `critique` already hold motion, and gives slices 9 and 10 something to extend and measure.

## Acceptance criteria (from docs/roadmap.md)

1. `packs/motion-reference/` passes `tenet validate` and holds an invented, labelled motion personality (point of view), motion tokens (easing, duration, stagger, spring, frame rate), 8 to 10 rules of which at least 4 are checkable, 3 motion components, one scene template and one storyboard flow.
2. The README states the pack is an invented example with no outside source, and that its rule patterns target Remotion and CSS source.
3. A fixture Remotion composition that uses only pack tokens reports no findings from `critique`.

## Decisions

| # | Decision | Reason |
|---|---|---|
| M1 | The pack lives at `packs/motion-reference/` with its design system at `packs/motion-reference/design-system/`, like the Stripe pack. No code changes. | ADR-0011: motion is a focus of packs, not a new file kind. |
| M2 | The personality is **Settled**: entrances decide quickly and then settle slowly with no overshoot; exits are shorter than entrances; one element is the focus of the eye at a time; the payoff is held still. The point of view says it is invented for this example, is no real brand's, and has no outside source. | There is no public source to reverse-engineer (ADR-0011). A clear, opinionated personality shows what a point of view buys over generic motion. |
| M3 | Token categories: `motion`, `stagger`, `spring`, `frame`, `color`, `type`. `motion`: `settle` `cubic-bezier(0.16, 1, 0.3, 1)`, `depart` `cubic-bezier(0.7, 0, 0.84, 0)`, `duration-fast` `160ms`, `duration-base` `400ms`, `duration-hero` `900ms`. `stagger`: `step` `60ms`, `limit` `5`. `spring`: `damping` `28`, `stiffness` `120`, `mass` `1`. `frame`: `fps` `30`, `landscape-width` `1920`, `landscape-height` `1080`. `color`: `ink`, `paper`, `accent`. `type`: `family` `"Inter", system-ui, sans-serif`, `weight-display` `500`. | Covers the roadmap's easing, duration, stagger, spring and frame-rate tokens. Colour and type give a motion piece something to draw with, so the colour rule is meaningful. |
| M4 | Token names must not match the pack's own `Forbid:` patterns (ADR-0010). Names avoid `ease`, `linear`, `bounce`, `elastic` and `delay`. | The `settle` and `depart` names, not `ease-out`, follow ADR-0010. |
| M5 | Nine rules, five checkable, four judgement. Checkable: `motion.no-builtin-easing`, `motion.springs-settle`, `motion.time-from-tokens`, `motion.stagger-from-tokens`, `color.tokens-only`. Judgement: `motion.one-focus`, `motion.reading-order`, `motion.hold-the-payoff`, `motion.exit-shorter-than-enter`. Each has a `Reason:` and an `Applies to:` line. | At least 4 checkable (criterion 1), both kinds of rule as ADR-0009 expects. |
| M6 | Patterns (Go RE2, no lookaround, matched per line). `motion.no-builtin-easing`: `(transition\|animation)[^;"]*\b(ease\|ease-in\|ease-out\|ease-in-out\|linear)\b\|\bEasing\.(linear\|ease\|quad\|cubic\|sin\|exp\|circle\|bounce\|elastic\|back)\b\|\bEasing\.bezier\(\s*[0-9.-]`. `motion.springs-settle`: `overshootClamping\s*:\s*false\|\bdamping\s*:\s*([0-9]\|1[0-9])\b`. `motion.time-from-tokens`: `(transition\|animation)[^;"]*\b[0-9]+(\.[0-9]+)?m?s\b\|\b(durationInFrames\|fps)\s*[=:]\s*\{?\s*[0-9]`. `motion.stagger-from-tokens`: `\b(delay\|stagger)[A-Za-z]*\s*[=:]\s*\{?\s*[0-9]`. `color.tokens-only`: the Stripe pack's pattern. Each `Fix:` names the token to use. | Targets CSS and Remotion, as criterion 2 says. A literal `Easing.bezier(0.1, ...)` is forbidden, so Remotion code spreads the pack curve: `Easing.bezier(...settle)`. |
| M7 | The Remotion convention: the project has a `tokens.ts` that exports the pack values (`settle` as a four-number array, `fps`, `durationBase` as a frame count helper, and so on). The fixture imports from it and never writes a raw time, curve or colour. The README documents the convention. | Gives the model one place to read token values in code, and lets the rules check the call sites. |
| M8 | Three components: `title-reveal` (words rise and settle with a stagger of `stagger.step`, at most `stagger.limit` steps), `number-counter` (counts up on `motion.settle` and holds the final value), `chart-draw-on` (a line or bars draw in once, then hold). Each cites that the personality is invented in its `Source:` line. | Primitives that most motion pieces need, each tied to a rule. |
| M9 | One scene template, `feature-showcase`: a title, one product visual and one proof number, 6 seconds at `frame.fps`. One flow, `launch-storyboard`: six scenes with duration, entrance, exit and pacing note, speeding in the middle and holding the last scene. | Criterion 1. The flow makes the pacing rules concrete. |
| M10 | Fixtures: `testdata/critique/motion-clean.tsx` (a Remotion composition that uses only pack tokens). The test file adds seeded-violation checks as in-memory edits of the fixture, one per checkable rule. | Criterion 3, and shows that every pattern is live. |
| M11 | Tests live in a new `internal/designsystem/motion_pack_test.go` and do not change `pack_test.go`. They assert: the pack loads with no problems; counts (9 rules, 5 checkable, 3 components, 1 template, 1 flow); the README contains `invented`, `no outside source`, `Remotion` and `CSS`; the pack ships only text files; the fixture has no findings; each seeded edit gives exactly one finding for the named rule; no pack file other than `rules.md` triggers any of the pack's own patterns (closes the ADR-0010 gap for this pack). | Repeats the slice 5 checks, plus the no-self-match test ADR-0010 logged as a candidate. |
| M12 | `testdata/project/`, the `tenet init` scaffold and the Stripe pack are unchanged. | Slice 2 and slice 4 keep those stable. |

## Content notes

- Judgement rules: `motion.one-focus` (only one element draws the eye at a time; others hold), `motion.reading-order` (staggered items enter in reading order), `motion.hold-the-payoff` (the key figure or message holds still for at least one second at the end of its scene), `motion.exit-shorter-than-enter` (an exit uses `duration-fast`, shorter than its entrance).
- Edge cases: a CSS custom-property declaration of a pack curve (`--motion-settle: cubic-bezier(...)`) is not flagged because `cubic-bezier` is not part of the easing pattern. A literal `Easing.bezier(0.16, ...)` in code is flagged on purpose. `ease` matches the whole word only, so `release` and `increase` do not match. A `fps` or `durationInFrames` set from a token expression (`fps={fps}`) does not match, since the character after the equals sign is not a digit.
- Known approximation: patterns are per line, so a value on the next line is missed (as in the Stripe pack).

## Architecture

```
packs/motion-reference/
├── README.md
└── design-system/
    ├── point-of-view.md
    ├── tokens.json
    ├── rules.md
    ├── components/{title-reveal,number-counter,chart-draw-on}.md
    ├── templates/feature-showcase.md
    └── flows/launch-storyboard.md
testdata/critique/motion-clean.tsx
internal/designsystem/motion_pack_test.go
```

## Testing

Per M11. `go test ./...`, `gofmt`, `golangci-lint` and `tenet validate --dir packs/motion-reference` must pass.
