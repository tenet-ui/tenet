# tenet

An MCP server that lets Claude users encode their own design system (point of view, tokens, components, templates, flows, rules) and delivers it to the AI at the right moment, so AI-built interfaces carry intent instead of generic "zombie UI". The Stripe design system, reverse-engineered from public sources, is the reference example.

## Language

**Design system**:
The whole harness a team uses to make interfaces: its point of view, rules, tokens, components, templates and flows. The object of design is the system, not a single screen.
_Avoid_: style guide, UI kit, component library (these name parts of it)

**Point of view**:
The brand intent behind the system: what the product stands for, who it is for, and what its users care about. The source every other rule is derived from.
_Avoid_: mission, values, vibe

**Rule**:
A single, checkable statement of how the system must or must not behave, with a reason. Rules are what a critique checks against.
_Avoid_: guideline, best practice, tip

**Token**:
A named design value (colour, space, radius, type scale, motion) that components reference instead of raw values.
_Avoid_: variable, constant, theme value

**Component**:
A reusable UI building block with defined props, states and usage rules.
_Avoid_: widget, element

**Template**:
A complete page or screen composed from components, showing how the system assembles a whole surface. Templates are first-class, not an afterthought to components.
_Avoid_: layout, scaffold, boilerplate

**Flow**:
A multi-step user journey across templates (for example, onboarding or checkout), with the states and transitions the system expects.
_Avoid_: user journey, wizard, process

**Harness**:
The mechanism that makes the AI follow the system: the files, the MCP tools that serve them, and the order in which they are loaded.
_Avoid_: prompt, system prompt, context dump

**Delivery**:
Serving the right slice of the design system to the AI at the moment it is needed (for example, the tokens for a component being built), instead of loading everything up front.
_Avoid_: retrieval, RAG, injection

**Critique**:
An editor pass over generated output that checks it against the rules, the point of view and the user journey, and returns specific fixes. A critique is not a yes/no gate.
_Avoid_: review, lint, validation

**Finding**:
One breach of one rule in one file, with the line, the rule's id, the matched text and the rule's fix. Critique returns findings; it does not pass or fail a file.
_Avoid_: error, violation, issue, warning

**Token-only output**:
A generated page with no critique findings against the pack and at least one named `var(--...)` value. It is a proxy: it does not check that each value is the right token for its role, and it counts a token declaration as a raw value (see experiments/consistency/REPORT.md).
_Avoid_: clean output, pure output

**Checkable rule**:
A valid rule with a `Forbid:` pattern. `critique` checks only checkable rules.
_Avoid_: automated rule, lint rule

**Judgement rule**:
A rule no pattern can check, such as whether a layout fits the user journey. The client reads it through the delivery tools and reasons over it (ADR-0009).
_Avoid_: soft rule, guideline

**Done vs good**:
The distinction that a generated interface can look finished while still failing the problem, the differentiation or the journey. Critique exists to close that gap.
_Avoid_: polish, finished

**Zombie UI**:
Interfaces built from generic, popular patterns with no intent behind them, the digital equivalent of the post-war zombie buildings. The failure mode tenet exists to prevent.
_Avoid_: generic UI, slop

**Reference system**:
A design system shipped with tenet as a worked example. The Stripe one is reverse-engineered, unofficial and built only from public sources; the motion one is invented and labelled as such.
_Avoid_: official Stripe system, Stripe pack

**Pack**:
A distributable, versioned folder of design-system files (tokens, rules, components, templates, flows, point of view) that a user can start from or fork.
_Avoid_: theme, template bundle, preset

**Linked rule**:
A rule whose `Applies to:` line names an item or token category, so that a delivery tool returns it with that item. A rule with no `Applies to:` line is valid but is never a linked rule.
_Avoid_: attached rule, related rule, tagged rule

**Intent**:
What the product and brand are trying to achieve for a user. Output quality is measured against intent, not against whether the output was AI-made.
_Avoid_: prompt, goal, request


**Motion**:
Anything in a made thing that changes over time: a UI transition, an animated explainer, a launch video. Motion is held to the same system as a static screen: it comes from tokens, follows rules and carries the point of view.
_Avoid_: animation effects, transitions (these name parts of it)

**Motion pack**:
A pack whose tokens, rules, components, templates and flows are about motion, such as easing, duration, stagger, scenes and storyboards. It uses the same files as any other pack; it is a pack with a motion focus, not a new kind of file.
_Avoid_: animation preset, motion theme

**Motion personality**:
How a brand moves: fast or slow, restrained or playful, whether it overshoots or settles. The motion part of the point of view, from which the motion tokens and rules are derived.
_Avoid_: animation style, feel

**Scene**:
A motion template: one self-contained stretch of a motion piece, such as a title card or a feature showcase, composed from motion components with set timings.
_Avoid_: slide, clip, shot

**Storyboard**:
A motion flow: the ordered scenes of a motion piece with the timing and transition between them and where the pacing speeds up or slows down.
_Avoid_: timeline, sequence, script

**Generic motion**:
The motion counterpart of zombie UI: built-in easing, one default duration, a fade-up on everything, a bounce on every element, with no personality behind it.
_Avoid_: default animation, slop
