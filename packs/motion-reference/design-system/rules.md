# Rules

## motion.no-builtin-easing
Motion uses the pack curves, `settle` and `depart`, never a built-in easing keyword or a literal curve. In Remotion, spread the token: `Easing.bezier(...settle)`.
Reason: Settled decides quickly and settles slowly; a built-in curve has no personality and a literal curve drifts from the pack.
Applies to: tokens:motion, component:title-reveal, component:number-counter, component:chart-draw-on
Forbid: (transition|animation)[^;]*\b(ease|ease-in|ease-out|ease-in-out|linear)\b|\bEasing\.(linear|ease|quad|cubic|sin|exp|circle|bounce|elastic|back|poly|step0|step1)\b|\bEasing\.bezier\(\s*-?\.?[0-9]
Fix: Use motion.settle for entrances or motion.depart for exits from design-system/tokens.json; in Remotion write Easing.bezier(...settle).

## motion.springs-settle
Springs set overshootClamping: true and use the spring tokens.
Reason: An overshooting spring bounces, and Settled never bounces; the eye should rest on the element, not follow it. The spring tokens are tuned to settle, and a hand-written low damping value drifts from them.
Applies to: tokens:spring, component:title-reveal, component:number-counter
Forbid: overshootClamping\s*:\s*false|\bdamping\s*:\s*([0-9]|1[0-9])\b
Fix: Use the spring tokens (damping, stiffness, mass) from design-system/tokens.json and set overshootClamping: true.

## motion.time-from-tokens
Durations and frame rates come from tokens: duration-fast, duration-base, duration-hero and frame.fps, not raw times or frame counts.
Reason: Raw timings break the rhythm of the piece; one scale keeps every scene on the same pulse.
Applies to: tokens:motion, tokens:frame, component:title-reveal, component:number-counter, component:chart-draw-on, template:feature-showcase
Forbid: (transition|animation)[^;]*\b[0-9]+(\.[0-9]+)?m?s\b|\b(durationInFrames|fps)\s*[=:]\s*\{?\s*[0-9]
Fix: Use motion.duration-fast, motion.duration-base or motion.duration-hero, and frame.fps, from design-system/tokens.json.

## motion.stagger-from-tokens
Stagger offsets come from stagger.step and run for at most stagger.limit steps.
Reason: A fixed step reads as one gesture; an arbitrary offset reads as noise.
Applies to: tokens:stagger, component:title-reveal
Forbid: \b(delay|stagger)[A-Za-z]*\s*[=:]\s*\{?\s*[0-9]
Fix: Use stagger.step from design-system/tokens.json and stop after stagger.limit steps.

## color.tokens-only
Components take colours from tokens, never from raw hex values or rgb() and hsl() functions.
Reason: A raw value drifts from the palette the moment one scene is edited alone.
Applies to: tokens:color, component:title-reveal, component:number-counter, component:chart-draw-on, template:feature-showcase
Forbid: #[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?)\(
Fix: Replace the raw colour with a colour token from design-system/tokens.json.

## motion.one-focus
Only one element draws the eye at a time; everything else holds still while it moves.
Reason: Settled leads the eye. Two things moving at once split attention and the piece loses its point.
Applies to: template:feature-showcase, flow:launch-storyboard

## motion.reading-order
Staggered items enter in reading order, first to last, left to right and top to bottom.
Reason: The motion should carry the viewer through the content in the order they would read it.
Applies to: component:title-reveal, template:feature-showcase

## motion.hold-the-payoff
The key figure or message holds still for at least one second at the end of its scene.
Reason: Settled ends on rest. A payoff that is still moving when the scene cuts is not read.
Applies to: component:number-counter, component:chart-draw-on, flow:launch-storyboard

## motion.exit-shorter-than-enter
An exit uses duration-fast and is shorter than the entrance it follows.
Reason: Settled decides quickly to leave; the viewer has already seen the content, so the exit should not make them wait.
Applies to: tokens:motion, flow:launch-storyboard
