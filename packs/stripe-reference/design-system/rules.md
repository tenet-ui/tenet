# Rules

## color.tokens-only
Components take colours from tokens, never from raw hex values or rgb() and hsl() functions.
Reason: A raw value drifts from the palette the moment one screen is edited alone (research section 1).
Applies to: tokens:color, component:button, component:text-field, template:payments-list
Forbid: #[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?)\(
Fix: Replace the raw colour with a colour token from design-system/tokens.json.

## spacing.tokens-only
Spacing between elements uses space tokens, not raw pixel values.
Reason: Off-grid spacing is the quickest way for a screen to look unplanned (research section 3, an 8px-based scale).
Applies to: tokens:space, component:button, component:text-field, template:payments-list
Forbid: \b(margin|padding|gap)\b[^;"]*\b\d+px
Fix: Replace the raw pixel value with a space token from design-system/tokens.json.

## radius.tokens-only
Corner radius comes from a radius token, not a raw pixel value.
Reason: The corner scale is small and fixed, so one raw value stands out (research section 3, radii of 2, 4, 6, 16 and 32px, and round).
Applies to: tokens:radius, component:button, component:text-field
Forbid: border-radius[^;"]*\b\d+px
Fix: Replace the raw pixel radius with a radius token from design-system/tokens.json.

## motion.no-builtin-easing
Transitions and animations use a cubic-bezier curve from the motion tokens, not a built-in easing keyword.
Reason: Built-in ease-in, ease-out and linear curves look generic next to a tuned curve (research section 4).
Applies to: tokens:motion, component:button
Forbid: (transition|animation)[^;"]*\b(ease-in|ease-out|linear)\b
Fix: Replace the easing keyword with the motion.standard token from design-system/tokens.json.

## button.one-primary
A view has at most one primary button; other actions use the secondary style.
Reason: Two primary actions compete, so the user cannot tell which one the screen expects (research section 7, checkout and onboarding patterns).
Applies to: component:button, template:payments-list

## empty-state.title-states-missing
An empty state's title states what is missing, as a short phrase ending in a period, and its action repeats the title's noun.
Reason: A title that promotes instead of stating the gap reads as marketing (research section 6, the Stripe Apps empty-state rules).
Applies to: component:empty-state, template:payments-list
