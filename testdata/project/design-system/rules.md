# Rules

## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
Applies to: tokens:color, component:button
Forbid: #[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?)\(
Fix: Replace the raw colour with a colour token from tokens.json.

## spacing.unit-multiples
Spacing between elements is a multiple of the space unit token.
Reason: Off-grid spacing is the quickest way for a screen to look unplanned.
Applies to: tokens:space, template:settings-page
Forbid: \b(margin|padding|gap)\b[^;"]*\b\d+px
Fix: Replace the raw pixel value with a space token from tokens.json, so spacing stays a multiple of the space unit.
