# Slice 5: Reference pack: unofficial Stripe-inspired system — design

Approved: 2026-10-18

Self-approved under the project's no-approval-gates rule (see CLAUDE.md, "Decide, don't ask"). Path: architectural in size (a new reference pack and a new fixture), but no Go interface changes: no new tool, no new exported type, no change to `designsystem` or `server` code.

## Goal

Ship `packs/stripe-reference/`, a worked design-system pack built from the public-source research in `docs/research/stripe-design-system.md`, so a user can see a complete system (point of view, tokens, rules, components, one template, one flow) and run `critique` against it. The pack is labelled as an unofficial, Stripe-inspired reverse engineering and ships no Stripe marks (ADR-0004).

## Acceptance criteria (from docs/roadmap.md)

1. A `packs/stripe-reference/` folder built from `docs/research/stripe-design-system.md`, with tokens, rules, components, one template and one flow, all citing their research source.
2. The pack ships no Stripe logos or wordmarks, and its README states it is an unofficial reverse engineering.
3. Running `critique` against a fixture page that uses only pack tokens reports no findings.

## Decisions

| # | Decision | Reason |
|---|---|---|
| R1 | The pack lives at `packs/stripe-reference/`, with its design system at `packs/stripe-reference/design-system/`. The `critique` `system` argument is `packs/stripe-reference`. | The slice 4 decision C8 already supports a folder holding its own `design-system/`, so no code changes are needed. |
| R2 | The pack holds: `point-of-view.md`; `tokens.json` with categories `color`, `space`, `radius`, `motion`, `type`; `rules.md` with four checkable rules and two judgement rules; components `button`, `text-field`, `empty-state`; template `payments-list`; flow `account-setup`. | Covers acceptance 1 and the research sections that the rules and items need. Three components are enough to show `Applies to:` links to more than one kind of item. |
| R3 | Token values are the pack's own choices, inspired by the hue families the research observed (indigo brand, blue-grey neutrals, navy-tinted shadows, a light heading weight). No observed hex value is copied. | ADR-0004: "a shipped theme should be derivative and clearly labelled as inspired by". Each token's observed source is named in the README (R9). |
| R4 | The pack uses no font file and no Söhne name. `type.family` is the system-safe stack `"Inter", system-ui, sans-serif`, and `type.weight-light` is `300`. The `type` category is present for the research's light display weight, but no rule checks it. | ADR-0004: no font files; Söhne is licensed to Stripe. |
| R5 | Four rules carry a `Forbid:` and `Fix:` pair (checkable): `color.tokens-only`, `spacing.tokens-only`, `radius.tokens-only`, `motion.no-builtin-easing`. Two rules are judgement only, with no `Forbid:` line: `button.one-primary` and `empty-state.title-states-missing`. | Shows both kinds of rule, as ADR-0009 expects. Judgement rules stay with the client (C4). |
| R6 | The four patterns (RE2, no lookaround): colour `#[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?)\(`; spacing `\b(margin|padding|gap)\b[^;"]*\b\d+px`; radius `border-radius[^;"]*\b\d+px`; motion `(transition|animation)[^;"]*\b(ease-in|ease-out|linear)\b`. Each `Fix:` names the token category to use. | RE2 has no lookahead, so the `ease-in` pattern also flags `ease-in-out`. That is accepted: `ease-in-out` is also a built-in keyword the rule forbids. The spacing approximation is inherited from slice 4 C17. |
| R7 | The fixture `testdata/critique/pack-clean.html` uses only `var(--<category>-<name>)` references for colours, spacing, radii and motion. It contains no raw hex, no `rgb()`/`hsl()`, no pixel value after `margin`, `padding`, `gap` or `border-radius`, and no built-in easing keyword. | Acceptance 3. The fixture is written against the pack's token names, so it reads as an example of tokens in use. |
| R8 | Tests check the pack against the fixture, not against an exact list of pack contents. They assert: the pack loads with no problems; it has one template and one flow; each checkable rule is valid; `Checkable` is 4; the fixture reports `No findings in ... Checked 4 rule(s).` | A check that the pack and the fixture are in step, without freezing every token value. |
| R9 | The README section names each token category's observed source (`docs/research/stripe-design-system.md`, section number), states that values are derivative, and states the Stripe-unofficial label. Each component, the template and the flow carry a `Source:` line citing their research section. Each rule's `Reason:` cites its research section in parentheses. | Acceptance 1 (citations) and acceptance 2 (the label). Token files cannot carry comments, so the README holds the token citations; the file format is unchanged. |
| R10 | A seeded-violation test: the fixture with one raw hex colour appended to its `<style>` block, run through `designsystem.Critique` with the pack's rules, gives exactly one finding, for `color.tokens-only`. | Shows the pack's patterns are not vacuous, so the "no findings" result means something. |
| R11 | No code changes to `internal/designsystem` or `internal/server`. The tool count stays at six, and the server test list is unchanged. | Keeps this slice to content plus tests. |
| R12 | `testdata/project/` and the `tenet init` scaffold are unchanged. | Slice 2 D26 and slice 4 C15 keep those stable. |

## Pack content (the design choices)

- **Point of view**: a calm, precise system for payment-adjacent products, written in the pack's own voice. It names the research as its reference and says it is not Stripe's.
- **Tokens** (`tokens.json`):
  - `color`: `ink`, `ink-soft`, `ink-quiet`, `paper`, `paper-subdued`, `border`, `brand`, `brand-hover`, `brand-soft`, `focus`, `danger`
  - `space`: `unit` 8, `xs` 4, `sm` 8, `md` 16, `lg` 24, `xl` 32
  - `radius`: `xs` 2, `sm` 4, `md` 6, `lg` 16, `round` 9999
  - `motion`: `ease-out` `cubic-bezier(0.25, 1, 0.5, 1)`, `duration-fast` `120ms`, `duration-base` `300ms`
  - `type`: `family` `"Inter", system-ui, sans-serif`, `weight-light` 300, `weight-bold` 500
- **Rules** (`rules.md`): as R5 and R6, each with a `Reason:` that cites the research section, and an `Applies to:` line naming components, the template, and token categories.
- **Components**: `button` (primary action; one per view), `text-field` (labelled input), `empty-state` (a title and one action).
- **Template**: `payments-list`, a filterable, exportable list of payments with a status column and a row-detail link, from research §7.
- **Flow**: `account-setup`, the six steps from research §7 (account, sandbox, checklist, business verification, public details, two-factor security), with the states and transitions for each.

## Architecture

```
packs/stripe-reference/
├── README.md                          # unofficial label, sources, token citations, how to run critique (R9)
└── design-system/
    ├── point-of-view.md
    ├── tokens.json
    ├── rules.md
    ├── components/
    │   ├── button.md
    │   ├── text-field.md
    │   └── empty-state.md
    ├── templates/
    │   └── payments-list.md
    └── flows/
        └── account-setup.md
testdata/critique/pack-clean.html      # fixture (R7)
internal/designsystem/pack_test.go     # R8, R10
internal/server/critique_test.go       # one added case: the pack and fixture, exact result text
README.md                              # one pointer to packs/stripe-reference (R9)
```

Dependency direction is unchanged. The pack is data; the tests read it through the existing `Load` and `Critique` functions.

## Data flow

Unchanged from slice 4. `critique` with `file = testdata/critique/pack-clean.html` and `system = packs/stripe-reference` runs `designsystem.Load` on the pack, then `Critique` over the fixture, then renders the result.

## Errors

No new error paths. A pack that stops loading cleanly is a test failure (R8), not a runtime error.

## Testing

- `internal/designsystem/pack_test.go`:
  - `Load("../../packs/stripe-reference")` returns no problems.
  - The pack has one template, one flow, three components, and six rules, of which four are valid with a `Forbid` pattern (`Checkable` is 4).
  - Every `Applies to:` ref resolves (already checked by `Load`, asserted explicitly).
  - `Critique` of the fixture against the pack's rules is empty.
  - The seeded-violation case (R10) gives one finding, for `color.tokens-only`.
- `internal/server/critique_test.go`: one case, the fixture against `packs/stripe-reference`, expects the exact text `No findings in testdata/critique/pack-clean.html against packs/stripe-reference. Checked 4 rule(s).`
- Full suite: `go test ./...`, the gofmt check, and golangci-lint (commands in CLAUDE.md). The pack's JSON is checked by `go run ./cmd/tenet validate --dir packs/stripe-reference` and must exit zero.

## Known limits (stated in the pack README)

- The pack is an unofficial, Stripe-inspired reverse engineering. Its token values are derivative, not Stripe's, and it is not endorsed by Stripe.
- The spacing pattern flags raw pixel values in spacing properties even where the rule allows them (inherited from slice 4 C17).
- The motion pattern flags `ease-in-out` as well as `ease-in`, `ease-out` and `linear`.
- Judgement rules (`button.one-primary`, `empty-state.title-states-missing`) are not checked by `critique`; they are read by the client through the delivery tools (ADR-0009).
- Sail, Stripe's internal design system, and Dashboard visuals were not observed; the pack does not claim to match them.
