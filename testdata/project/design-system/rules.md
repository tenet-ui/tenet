# Rules

## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
Applies to: tokens:color, component:button

## spacing.unit-multiples
Spacing between elements is a multiple of the space unit token.
Reason: Off-grid spacing is the quickest way for a screen to look unplanned.
Applies to: tokens:space, template:settings-page
