# Motion reference pack

**Invented example.** This pack is an invented motion personality called Settled. It belongs to no real brand and has no outside source: nothing in it is reverse-engineered or copied from a product. It ships text files only.

It is a worked example of a design system for motion: a point of view, tokens (easing, duration, stagger, spring, frame rate), nine rules, three components, a scene template and a storyboard flow. `critique` checks source against its rules, and the delivery tools hand the rest to your client.

## Tokens

| Category | Tokens |
|---|---|
| motion | `settle`, `depart`, `duration-fast` (160ms), `duration-base` (400ms), `duration-hero` (900ms) |
| stagger | `step` (60ms), `limit` (5) |
| spring | `damping`, `stiffness`, `mass` |
| frame | `fps` (30), `landscape-width`, `landscape-height` |
| color | `ink`, `paper`, `accent` |
| type | `family`, `weight-display` |

Token names avoid the words the pack's own patterns forbid, so the pack cannot trigger its own rules.

## What the patterns check

The checkable rules target Remotion and CSS source. They check one line at a time and check text, not structure:

- Built-in easing keywords (the CSS keywords and the built-in Remotion easing functions) and a bezier written with literal numbers are flagged. Spread the pack curve instead, as the fixture does.
- Spring configs that overshoot or under-damp, raw times in a CSS `transition`, raw frame counts, raw stagger offsets and raw colours are flagged.
- A CSS custom property that declares a pack curve is not flagged.

A value written on the next line is missed. The judgement rules are not checked by `critique`; the client reads them through the delivery tools.

## Known limits

- Patterns are per line and text-based, so comments and keyframe names containing the keywords can match.
- `transition-delay` and `animation-delay` can produce two findings.
- Not caught: a default `spring({ frame, fps })` with no config, raw frame counts in `<Sequence from={45}>` or `interpolate(frame, [0, 30], ...)`, and `frame - i * 4` staggers.
- The colour rule does not match `oklch()` values, which are this pack's colour format.
- `tokens.ts` and `tokens.json` declare the values, so do not run `critique` on `tokens.ts`: it would match its own declarations. Exempting token declarations is a logged follow-up.

## The `tokens.ts` convention

A Remotion project using this pack keeps one `tokens.ts` that exports the pack values: `settle` as a four-number array, `fps`, `ms()`, a helper that turns a duration token into a frame count, and the stagger, spring and colour values. Compositions import from it and never write a raw time, curve or colour. The fixture `testdata/critique/motion-clean.tsx` follows this and reports no findings.

## Validate and critique

Validate the pack from the repository root:

    go run ./cmd/tenet validate --dir packs/motion-reference

Ask your MCP client to call `critique` with `file` set to a Remotion or CSS source file and `system` set to `packs/motion-reference`.
