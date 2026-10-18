---
status: accepted
---

# Reference packs use derivative values, and their names must not trigger their own rules

A reference pack (`packs/stripe-reference/`) is built from public observations (ADR-0004), so its token values are the pack's own choices. No observed hex value is copied. Two values may reuse the research's observed values, because they describe structure rather than brand colour: the easing curve and the radius scale. The pack README discloses this.

A rule's `Forbid:` pattern runs over every source file, and a pack's token names appear in its own files, so a token name can trigger the pack's own rules. The motion token is named `standard`, not `ease-out`: `var(--motion-ease-out)` inside a transition matches `motion.no-builtin-easing`. Pack authors choose token names that their own patterns cannot match.

The trade-off is that the reused curve and radii are not wholly the pack's own, and the no-self-match rule is enforced by review, not by a test. A test that runs the pack's own rules over its own token names would close that gap; it is logged as a candidate under the Slice 5 retro, not built here.