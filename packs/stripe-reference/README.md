# Stripe-inspired reference pack

**Unofficial.** This pack is a reverse engineering of Stripe's public design system, built only from publicly observable sources. It is not a Stripe product, it is not endorsed by Stripe, and it ships no Stripe logos, wordmarks, font files or other brand assets. Every part is labelled "inspired by".

It is a worked example: a complete design system (point of view, tokens, rules, components, a template and a flow) that `critique` can check.

## Sources

Every observation traces to [docs/research/stripe-design-system.md](../../docs/research/stripe-design-system.md). That file lists its public sources, the date accessed (2026-10-18), and a confidence level for each observation.

Token values are this pack's own choices, inspired by the hue families the research observed. No observed hex value is copied. The easing curve and the radius scale reuse the research's observed values, because they describe structure rather than brand colour.

| Token category | Based on research section |
|---|---|
| color | 1. Colour palette and tokens |
| type | 2. Typography |
| space | 3. Spacing, radius, shadows, borders |
| radius | 3. Spacing, radius, shadows, borders |
| motion | 4. Motion |

Each item file cites its own sections in a `Source:` line.

## Check a page

Validate the pack from the repository root:

    go run ./cmd/tenet validate --dir packs/stripe-reference

Ask your MCP client to call `critique` with `file` set to a UI source file and `system` set to `packs/stripe-reference`. The fixture `testdata/critique/pack-clean.html` uses only pack tokens and reports no findings.

## Known limits

- Patterns check one line at a time, and check text, not structure.
- The spacing pattern flags raw pixel values even where the rule allows them.
- The motion pattern flags `ease-in-out` as well as `ease-in`, `ease-out` and `linear`.
- The judgement rules `button.one-primary` and `empty-state.title-states-missing` are not checked by `critique`; the client reads them through the delivery tools.
- Sail, Stripe's internal design system, and Dashboard visuals were not observed. This pack does not claim to match them.
