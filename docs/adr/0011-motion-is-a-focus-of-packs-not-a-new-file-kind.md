---
status: accepted
---

# Motion is a focus of packs, not a new file kind, and the first target is Remotion

Motion graphics use the pack format as it is. A motion pack is a pack whose tokens (easing curves, durations, stagger delays, spring settings, frame rate, aspect ratios), rules, components (text reveal, counter, chart draw-on), templates (scenes) and flows (storyboards) are about motion. No new file type, tool or loader code is needed for delivery: `get_tokens`, `get_component`, `get_template` and `get_flow` already serve them (ADR-0006). The alternative, a separate motion schema with timeline and keyframe files, would double the format and the tools before we know what a motion pack needs.

The first target is code-based motion in Remotion (React that renders to video), with plain CSS second. Source files can be checked by `critique`'s patterns (ADR-0009), and Claude writes this code well. Visual tools (After Effects, Lottie authoring) have no source to check and are out of scope. GSAP and Framer Motion follow later, using the same token and rule format.

Two limits are accepted. First, `critique` reads source, not rendered video, so whether a piece feels right (pacing, hierarchy, whether motion supports the message) stays a judgement rule for the client. A frame-sampling tool that lets the client look at rendered frames is a later decision, not built now. Second, a motion personality has no public source to reverse-engineer as Stripe did (ADR-0004), so the first motion reference pack is an invented, labelled example. The existing token-naming hazard in ADR-0010 applies: motion token names must not match the pack's own `Forbid:` patterns.
