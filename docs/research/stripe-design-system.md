# Stripe design system: public-source research

Accessed 2026-10-18. Public sources only. This file describes Stripe's brand marks (logo, wordmark, parallelogram) in words and does not reproduce them.

## Summary

1. stripe.com runs on a token layer with the prefix `hds-`, defined in one CSS bundle. It has core ramps (brand, neutral, lemon, magenta, orange, ruby, success, error), semantic aliases (text, surface, action, accent), and `.hds-mode--light` / `.hds-mode--dark` modes.
2. The current core brand purple is `#533afd` (`brand-600`). The older "blurple" `#635bff` still appears in gradients, the focus ring, and the legacy newsroom theme, with navy `#0a2540` as the legacy dark.
3. Type is one variable face, `sohne-var` (Söhne, weight range 1–1000), with fallbacks `"SF Pro Display", sans-serif`. Code uses `SourceCodePro` Medium. Headings use weight 300 and tight negative tracking. The scale steps up at 640px and 940px.
4. Spacing uses an 8px-based `space-core-N` scale (N/100 × 8px). Radii run 2, 4, 6, 16, 32px, plus round. Shadows are two layers tinted blue (`rgba(0,55,112,…)`). Borders are 1px in "quiet" neutrals.
5. Motion favours ease-out curves, chiefly `cubic-bezier(.25,1,.5,1)` at 0.3s, with 62 `prefers-reduced-motion: no-preference` gates. A 2017 Stripe post says to animate only `transform` and `opacity` and to keep most animations under 500ms.
6. Stripe documents its public UI customisation surfaces in detail: the Elements Appearance API (3 themes, about 70 variables, rules on public class names), Connect embedded components appearance, Checkout branding settings, and Stripe Apps tokens and components (`@stripe/ui-extension-sdk`).
7. Stripe Apps docs publish writing rules for empty states, plus onboarding, progress, and loading patterns. This is the closest thing to a public product writing guide. No general Stripe voice or style guide was found.
8. Katie Dill's 2026 talk (checked against a third-party Whisper transcript) confirms all the briefed claims. They include optimism as a value, an MCP over design docs that gave poor results, then a CLI with full templates and flows, "Pepsi bubbling", the marbled *Built to Grow* cover, 56 animation iterations, and "scale intent".
9. The internal design system is called **Sail**. Stripe's job listings and How I AI coverage of Protodash mention it, but its components, tokens, Figma files and the design-system CLI are not public.
10. Legal boundary: Stripe's Mark Usage Terms define "Marks" to include trade dress and design elements, and they forbid altering the marks or implying endorsement. Logo and badge kits are downloadable from the newsroom, subject to those terms.

## Sources

All accessed 2026-10-18.

- stripe.com homepage HTML: https://stripe.com/
- stripe.com token CSS bundle (hashed name, will rotate): https://b.stripecdn.com/mkt-ssr-statics/assets/_next/static/css/8035801299d8c712.css
- stripe.com homepage CSS bundle: https://b.stripecdn.com/mkt-ssr-statics/assets/_next/static/css/a0080d082229394d.css
- stripe.com navigation CSS bundle: https://b.stripecdn.com/mkt-ssr-statics/assets/_next/static/css/e4746273d30fe790.css
- Newsroom / brand assets (legacy theme CSS inline): https://stripe.com/newsroom/information (redirect target of https://stripe.com/newsroom/brand-assets)
- Mark Usage Terms: https://stripe.com/legal/marks
- Elements Appearance API: https://docs.stripe.com/elements/appearance-api.md?api-integration=paymentintents
- Payment Element: https://docs.stripe.com/payments/payment-element.md
- Checkout overview: https://docs.stripe.com/payments/checkout.md
- How Checkout works (hosted): https://docs.stripe.com/payments/checkout/how-checkout-works.md?payment-ui=stripe-hosted
- Checkout appearance (hosted): https://docs.stripe.com/payments/checkout/customization/appearance.md?payment-ui=stripe-hosted
- Connect embedded components appearance: https://docs.stripe.com/connect/embedded-appearance-options.md
- Stripe-hosted Connect onboarding: https://docs.stripe.com/connect/hosted-onboarding.md
- Account setup: https://docs.stripe.com/get-started/account/activate.md
- Web Dashboard basics: https://docs.stripe.com/dashboard/basics.md
- Stripe Apps design: https://docs.stripe.com/stripe-apps/design
- Stripe Apps style: https://docs.stripe.com/stripe-apps/style
- Stripe Apps patterns: https://docs.stripe.com/stripe-apps/patterns.md
- Stripe Apps onboarding pattern: https://docs.stripe.com/stripe-apps/patterns/onboarding-experience.md
- Stripe Apps empty state pattern: https://docs.stripe.com/stripe-apps/patterns/empty-state.md
- Stripe Apps progress stepping: https://docs.stripe.com/stripe-apps/patterns/progress-stepping.md
- npm `@stripe/ui-extension-sdk` 9.4.0 (MIT): https://registry.npmjs.org/@stripe/ui-extension-sdk/-/ui-extension-sdk-9.4.0.tgz
- npm `@stripe/stripe-js` 10.1.0 (MIT): https://registry.npmjs.org/@stripe/stripe-js/-/stripe-js-10.1.0.tgz
- npm `@stripe/react-stripe-js` 7.0.0 (MIT): https://registry.npmjs.org/@stripe/react-stripe-js/-/react-stripe-js-7.0.0.tgz
- GitHub: https://github.com/stripe/react-stripe-js, https://github.com/stripe/stripe-js, https://github.com/stripe/stripe-apps, https://github.com/stripe/elements-examples, https://github.com/markdoc/markdoc
- Blog, "Designing accessible color systems" (2019): https://stripe.com/blog/accessible-color-systems
- Blog, "Connect: behind the front-end experience" (2017): https://stripe.com/blog/connect-front-end-experience
- Stripe Press: https://press.stripe.com/
- Katie Dill, Lenny and Friends Summit talk (video): https://www.youtube.com/watch?v=GLvFTMtw4Jk. Cited as "Katie Dill, Lenny's Podcast / Lenny and Friends Summit transcript, YouTube GLvFTMtw4Jk".
- Third-party Whisper transcript of that talk (unofficial): https://hraness.com/reading/raise-the-ceiling-how-to-scale-intent-quality-and-artistry
- Summit index: https://www.lennysnewsletter.com/p/all-of-the-lenny-and-friends-summit
- Lenny's Podcast, "Building beautiful products with Stripe's Head of Design" (2023-10-15): https://www.lennysnewsletter.com/p/building-beautiful-products-with
- How I AI, Owen Williams on Protodash (2026-05-04): https://www.lennysnewsletter.com/p/the-internal-ai-tool-thats-transforming (video https://www.youtube.com/watch?v=hQFEAZK__q0)
- Secondary coverage of Protodash: https://www.chatprd.ai/how-i-ai/stripe-owen-williams-on-buildling-internal-prototyping-studio

## Open questions

- Which `hds` value applies at which breakpoint? The base, then the 640px override, then the 940px override was confirmed for the `heading-xxl` ramp only. The other three-value tokens are assumed to follow the same order.
- What does `hds` stand for? The CSS never expands it, and it is not confirmed to be Sail. Sail powers the product (Dashboard), and `hds` powers marketing; whether they share tokens is unknown.
- What are the exact default variable values of the Elements `stripe`, `night` and `flat` themes? The docs show only an example object (`colorPrimary: '#0570de'`, …), not defaults.
- Dashboard UI tokens (Sail) cannot be observed without a logged-in Dashboard session, which this research did not use.
- The homepage hero "wave" is a script-driven graphic with `.webp` fallbacks (`wave-fallback-*`). Its shader or animation parameters were not extracted from JS.
- The Stripe Apps colour and typography token tables (pixel and hex values) are rendered client-side and missing from the `.md` docs. Only the token names are known, from the SDK types.
- The Figma toolkit at `@stripedesign` on Figma Community was referenced by older Apps docs (per search results) and returned HTTP 403 to fetch. Its current status and licence are unverified.
- Katie Dill's talk quotes are verified only against an unofficial auto-transcript. Exact wording should be checked against the video.

---

## 1. Colour palette and tokens

### Core ramps (light mode, `:root` / `.hds-mode--light`)

- The token prefix is `--hds-color-*`, in three layers: `core-*` (raw ramps), `util-*` (utility aliases), and semantic groups (`text`, `surface`, `action`, `accent`, `button`, `input`, `shadow`) — [source: https://b.stripecdn.com/mkt-ssr-statics/assets/_next/static/css/8035801299d8c712.css] (confidence: observed)
- Brand ramp: 25 `#f5f5ff`, 50 `#e8e9ff`, 75 `#e2e4ff`, 100 `#d6d9fc`, 200 `#b9b9f9`, 300 `#9a9afe`, 400 `#7f7dfc`, 500 `#665efd`, **600 `#533afd`**, 700 `#4032c8`, 800 `#2e2b8c`, 900/925 `#1c1e54`, 950 `#161741`, 975 `#0f1137` — [source: same CSS] (confidence: observed)
- Brand alpha variants: 25A `#acacff1f`, 50A `#a3a7ff40`, 100A `#7680f54d`, 400A `#5452fbbf` — [source: same CSS] (confidence: observed)
- Dark-mode brand ramp (`brandDark`): 25 `#f6f7ff`, 50 `#e4eaff`, 75 `#ccdaff`, 100 `#c3d3ff`, 200 `#a8bfff`, 300 `#92adff`, 400 `#7389ff`, 500 `#5d64fe`, 600 `#533afd`, 700 `#362baa`, 800 `#2c2484`, 900 `#222069`, 925 `#1c1b5a`, 950 `#191a51`, 975 `#171055` — [source: same CSS] (confidence: observed)
- Neutral ramp (blue-tinted slate): 0 `#fff`, 25 `#f8fafd`, 50 `#e5edf5`, 100 `#d4dee9`, 200 `#bac8da`, 300 `#95a4ba`, 400 `#7d8ba4`, 500 `#64748d`, 600 `#50617a`, 700 `#3c4f69`, 800 `#273951`, 900 `#1a2c44`, 950 `#11273e`, 975 `#0d253d`, 990 `#061b31` — [source: same CSS] (confidence: observed)
- Neutral dark ramp (navy): 25 `#f2f7fe`, 50 `#e3ecf7`, 100 `#d4deef`, 200 `#c0cee6`, 300 `#a3b5d6`, 400 `#839bc8`, 500 `#6480b2`, 600 `#45639d`, 700 `#273f73`, 800 `#23356e`, 900 `#182659`, 950 `#122054`, 975 `#101d4e`, 990 `#0d1738` — [source: same CSS] (confidence: observed)
- Accent ramps: lemon 25 `#fff2d8` / 100 `#ffe1a3` / 200 `#f9b900` / 300 `#e8a30b` / 500 `#9b6829` / 800 `#483414`; magenta 50 `#ffe6f5` / 100 `#ffd7ef` / 350 `#f44bcc` / 500 `#cf2cab` / 600 `#a51d85` / 800 `#4f2055`; orange 50 `#ffe5da` / 100 `#ffd8c6` / 350 `#ff6118` / 500 `#d04900` / 600 `#ab3500` / 800 `#56281b`; ruby 50 `#fee8eb` / 100 `#fed9de` / 400 `#ea2261` / 500 `#e2225f` / 600 `#b51145` / 800 `#5d1c3c` — [source: same CSS] (confidence: observed)
- Status ramps: success 100 `#b6f2c7` / 400 `#00b261` / 600 `#006f3a`; error 100 `#feb9ac` / 300 `#f4745c` / 400 `#f3432a` / 500 `#d8351e` / 600 `#a01400` / 700 `#721e11` — [source: same CSS] (confidence: observed)

### Semantic aliases

- Light-mode text: `text-solid` = neutral-990 `#061b31`, `text-soft` = neutral-600, `text-subdued` = neutral-500, `text-quiet` = neutral-400. `body` text colour is `text-soft` — [source: same CSS] (confidence: observed)
- Dark mode (`.hds-mode--dark`): `text-solid` = neutral-0 (white), `text-soft` = neutralDark-300, `text-subdued` = neutralDark-400, `text-quiet` = neutralDark-500 — [source: same CSS] (confidence: observed)
- Surfaces: `surface-bg-quiet` = white, `surface-bg-subdued` = neutral-25 `#f8fafd` (dark: neutralDark-950 `#122054`), `surface-border-quiet` = neutral-50 `#e5edf5` (dark: neutralDark-900) — [source: same CSS] (confidence: observed)
- Primary action: `action-bg-solid` = brand-600 `#533afd`, hover = brand-700 `#4032c8`. Quiet bg = brand-50, subdued bg = brand-75. Dark mode solid = brandDark-600, hover = brandDark-500 — [source: same CSS] (confidence: observed)
- Focus: outline `3px solid var(--hds-color-action-focus-outerSoft)`, where outerSoft = `#7b7eff96` or brand-400A. Offset 3px, 1px on buttons. Error focus `#d8280f80` — [source: same CSS] (confidence: observed)
- Primary buttons map `button-primary-bg` → `action-bg-solid` and hover → `action-bg-solidHover`; disabled uses `neutral-50A` — [source: same CSS] (confidence: observed)
- Accent colour modes (`accentColorMode-{default,lemon,magenta,orange,ruby}`) each define `icon-solid`, `icon-gradientStart/Middle/End`, `surface-subdued`, `border-quiet`. The explicit gradient triplets are lemon `#ffd552 → #ffaf2d → #ff9014`, magenta `#f98bf9 → #f96bee → #b262f9`, orange `#fe8c2d → #fd6252 → #fd5d7c`, and ruby starting `#fd7184` — [source: same CSS] (confidence: observed)

### Gradient and "blurple" usage

- Accent gradient stops: `--accent-gradient-color-stop-1: #bdb4ff`, `-2: #643afd`, `-3: #533afd` — [source: same CSS] (confidence: observed)
- Stat numerals use a text-clip gradient `linear-gradient(270deg, #ffd601, #ee30fb, #635bff)`, with stop positions shifted per stat (`--gradient-stat-one` … `-four`) so that a row of stats reads as one sweep from purple to yellow — [source: https://b.stripecdn.com/mkt-ssr-statics/assets/_next/static/css/a0080d082229394d.css] (confidence: observed for values; inferred for the "one sweep" intent, based on the shifted stops)
- Other observed homepage gradients: `linear-gradient(90deg,#7232f1 3.13%,#fb76fa 50%,#ffcf5e)`, `linear-gradient(203deg,#ffa319,#fd6252,#fd5b86)`, `linear-gradient(288.31deg,#0d1738,#4032c8)`, `linear-gradient(0deg,#7500fb,#ff39db)`, and translucent mixes of `rgba(83,58,253,.08)` / `rgba(255,140,108,.8)` / `rgba(218,75,254,.8)` over `#e3ecf7` — [source: same homepage CSS] (confidence: observed)
- `#635bff` (the historic "blurple") survives in the new system as `--hds-canary-color-border-focus` and in gradients, while the core brand-600 is the more saturated `#533afd` — [source: token CSS] (confidence: observed). This suggests a recent brand-colour shift from `#635bff` to `#533afd` — (confidence: inferred, from the coexistence of both values and from the legacy theme below)
- The legacy theme system, still served on the newsroom page, defines `.theme--Dark` bg `#0a2540`, `.theme--Light` bg `#f6f9fc`, `.theme--SemiDark` `#0d2e4f`, and text `#425466` and `#adbdcc`. Accents: Blurple `#635bff` / `#7a73ff` / `#6061f6`, Cyan `#00d4ff` / `#02bcf5`, Blue `#0073e6`, Pink `#f363f3`, Purple `#96f`, Raspberry `#ff5996`, Orange `#ff7600`, Green `#15be53`, Teal `#00c4c4`, Slate `#0a2540` — [source: https://stripe.com/newsroom/information] (confidence: observed)
- Stripe documents that its colour system is built in CIELAB so that hues share one lightness curve. Any two colours at least 5 levels apart pass small-text contrast, and 4 levels apart pass for icons and large text — [source: https://stripe.com/blog/accessible-color-systems] (confidence: documented)
- Katie Dill says optimism, a brand value, shows up in Stripe's colours, writing and subjects, and in products made for entrepreneurs — [source: Katie Dill, Lenny's Podcast / Lenny and Friends Summit transcript, YouTube GLvFTMtw4Jk; transcript https://hraness.com/reading/raise-the-ceiling-how-to-scale-intent-quality-and-artistry] (confidence: documented)

## 2. Typography

- Primary family token: `--hds-font-family: "sohne-var","SF Pro Display",sans-serif` — [source: token CSS] (confidence: observed)
- `@font-face` declares `sohne-var` from one variable file, `Sohne.cb178166.woff2`, as `format("woff2-variations")` with `font-weight: 1 1000` (one copy uses `font-display: block`, one `swap`) — [source: token CSS] (confidence: observed). Söhne is a commercial typeface by Klim Type Foundry and needs its own licence; Stripe's file must not be reused — (confidence: inferred, from the font name; licence terms were not checked here)
- Code family: `--hds-font-family-code: "SourceCodePro","SFMono-Regular",monospace`, with Source Code Pro Medium (500) self-hosted — [source: token CSS] (confidence: observed)
- Weights: `--hds-font-weight-normal` 300 or 400 and `--hds-font-weight-bold` 400 or 500 (two definitions each, likely per context or mode). A `weight-medium` token is also declared. Buttons use `weight-bold` — [source: token CSS] (confidence: observed)
- Every heading token (`heading-xxs` … `heading-xxl`, `heading-hero-*`) uses weight 300 except `xs` and `xxs` (400). That makes the system light-weight display type — [source: token CSS] (confidence: observed)
- Body text tokens `text-xxs` … `text-xxl` all use weight 300 — [source: token CSS] (confidence: observed)
- The heading scale has three values per token, at base, then `@media (min-width:640px)`, then `@media (min-width:940px)` — [source: token CSS] (confidence: observed for `heading-xxl`; inferred for the rest from the same declaration order)
  - `heading-xxl`: 2.125rem → 3rem → 3.5rem, line-height 1.03, tracking −0.02em → −0.025em at 940px
  - `heading-xl`: 1.75rem → 2.125rem → 3rem, line-height 1.07 / 1.05 / 1.03, tracking −0.01 / −0.02em
  - `heading-lg`: 1.375 / 1.75 / 2rem, line-height 1.2 / 1.1 / 1.07
  - `heading-md`: 1.25 / 1.375 / 1.625rem, line-height 1.2 / 1.12 / 1.1
  - `heading-sm`: 1.125 / 1.25 / 1.375rem, line-height 1.25 / 1.12 / 1.1
  - `heading-xs`: 1rem / 1.2, weight 400; `heading-xxs`: 0.875rem / 1.2, weight 400
  - `heading-hero-lg` up to 2.5rem, `hero-md` up to 2.25rem, `hero-sm` up to 2rem, line-height around 1.1
- Text scale: `text-xxs` 0.75rem/1.45; `text-xs` 0.75→0.875rem/1.4–1.45; `text-sm` 0.875rem/1.4; `text-md` 1rem/1.4; `text-lg` 1→1.125rem/1.35–1.4; `text-xl` 1.125→1.25rem/1.4; `text-xxl` 1.75→3rem/1–1.07 — [source: token CSS] (confidence: observed)
- Further type tokens exist for `input-label`, `input-text`, `input-description` (0.875rem/1.3/300), `input-groupHeading`, `quote-md`, `quoteAttribution-md` — [source: token CSS] (confidence: observed)
- OpenType: `font-feature-settings: "ss01" on` (stylistic set 1) in 3 rules, `"tnum"` for tabular numbers, and `"sups"` for footnote markers. The hero uses a `tabular-nums--tight` class for animated figures — [source: token CSS, homepage HTML https://stripe.com/] (confidence: observed)
- Buttons: 1rem, line-height 1, letter-spacing 0, block padding 15.5px top and 16.5px bottom (optical centring) — [source: token CSS] (confidence: observed for values; inferred for the "optical centring" intent)
- Elements default font: none documented. The Appearance API example uses `fontFamily: 'Ideal Sans, system-ui, sans-serif'`, and the docs advise at least 16px for mobile inputs — [source: https://docs.stripe.com/elements/appearance-api.md?api-integration=paymentintents] (confidence: documented)
- Stripe Apps typography is limited to the presets `body`, `bodyEmphasized`, `caption`, `heading`, `kicker`, `lead`, `subheading`, `subtitle`, `title`, plus `fontWeight` `regular | semibold | bold` and `fontFamily` `ui | monospace`. Arbitrary fonts are not allowed — [source: `@stripe/ui-extension-sdk` 9.4.0 `ui/index.d.ts`; https://docs.stripe.com/stripe-apps/style] (confidence: observed)

## 3. Spacing, radius, shadows, borders

- Spacing scale `--hds-space-core-N`: N/100 × 8px, i.e. 25=2px, 50=4px, 75=6px, 100=8px, 150=12px, 200=16px, 250=20px, 300=24px, 350=28px, 400=32px, 450=36px, 500=40px, 550=44px, 600=48px, 700=56px, 800=64px, 900=72px, 1000=80px … 2500=200px (8px steps above 1000), plus `core-1` = 1px — [source: token CSS] (confidence: observed)
- Layout: 4 / 8 / 12 columns (`--hds-space-layout-columns`), content max width 1264px, gap and content margin 16px, page margin 0 or 16px. Breakpoints used are 640px, 940px and 1264px (plus rare 600 and 840px) — [source: token CSS, homepage CSS] (confidence: observed)
- Button height `space-core-550` (44px) or `space-core-600` (48px), button border 1px — [source: token CSS] (confidence: observed)
- Radii: `radius-none` 0, `xs` 2px, `sm` 4px, `md` 6px, `lg` 16px, `xl` 32px, `round` 99999px. Usage counts in the CSS: `md` 114×, `sm` 51×, `xs` 17×, so 6px is the default corner — [source: token CSS] (confidence: observed for counts; inferred for "default")
- Button radius: `button-radius-lg` = `radius-sm` (4px), `button-radius-sm` = `radius-xs` (2px). Marketing buttons are nearly square, not pill-shaped — [source: token CSS] (confidence: observed)
- Shadows are always two stacked layers (top + bottom), tinted navy rather than black:
  - `xs`: 0 2px 10px `rgba(0,55,112,.06)`, 0 1px 4px `rgba(0,59,137,.04)`
  - `sm`: 0 5px 14px `rgba(0,55,112,.08)`, 0 2px 8px `rgba(0,59,137,.05)`
  - `md`: 0 6px 22px `rgba(0,55,112,.1)`, 0 4px 8px `rgba(0,59,137,.02)`
  - `lg`: 0 15px 40px −2px `rgba(0,55,112,.1)`, 0 5px 20px −2px `rgba(0,59,137,.04)`
  - `xl`: 0 20px 80px −16px `rgba(0,55,112,.14)`, 0 10px 60px −16px `rgba(0,59,137,.06)`
  - Dark mode swaps the shadow colours to `#0d173826` / `#0d173833` — [source: token CSS] (confidence: observed)
- Borders: the common pattern is `1px solid var(--hds-color-surface-border-quiet)` (neutral-50 `#e5edf5`), with `input-border-default` for inputs and a dashed `1px dashed` variant for placeholders — [source: token CSS] (confidence: observed)
- Stripe Apps spacing tokens: `xxsmall` 2px, `xsmall` 4px, `small` 8px, `medium` 16px, `large` 24px, `xlarge` 32px, `xxlarge` 48px. Radius tokens: `none | xsmall | small | medium | large | rounded`. Shadow tokens: `none | base | top | hover | focus`. Border colours: `neutral | critical` — [source: https://docs.stripe.com/stripe-apps/style; SDK types] (confidence: documented / observed)
- Connect embedded component defaults: `spacingUnit` 8px, `borderRadius` 15px, `buttonBorderRadius` 4px, `formBorderRadius` 6px, `badgeBorderRadius` 4px, `overlayBorderRadius` 8px, `colorBorder` `#D7D7D7` — [source: https://docs.stripe.com/connect/embedded-appearance-options.md] (confidence: documented)

## 4. Motion

- The dominant easing is `cubic-bezier(.25,1,.5,1)` (an ease-out quart-like curve), used 44× and typically at `.3s`, for `background-color`, `color`, `outline`, `border` and `fill` on buttons and links — [source: token CSS, homepage CSS] (confidence: observed)
- Other curves seen: `cubic-bezier(.4,0,.2,1)` (8×), `(0,0,.2,1)` (5×), `(.33,1,.68,1)` (5×), `(.16,1,.3,1)` (4×, used for `transform 1s`), `(.3,0,.2,1)` (opacity .4s), `(.65,.05,.36,1)`, and `(.45,.05,.55,.95)` — [source: all homepage CSS bundles] (confidence: observed)
- Named motion variables: `--card-duration: 800ms` with `--card-ease: cubic-bezier(0.165,0.84,0.44,1)` (transform and clip-path on cards); `--navigation-duration: 240ms`, `--navigation-duration-slow: 300ms`, `--navigation-easing: cubic-bezier(0.45,0.05,0.55,0.95)`; `--hds-accordion-duration: 0.36s` with `--hds-accordion-ease: cubic-bezier(0.65,0.05,0.36,1)` — [source: homepage, nav and token CSS] (confidence: observed)
- Quick UI feedback uses `.12s`–`.15s ease` for colour and opacity — [source: CSS bundles] (confidence: observed)
- Keyframe animations: `animatedgradient 1.5s ease-out infinite`, `gradient-border-input-shimmer 1s cubic-bezier(.39,.58,.57,1) 1.85s`, `book-of-the-week-fade-in .5s cubic-bezier(.33,1,.68,1)`, and scroll-driven `detect-scroll` timelines — [source: CSS bundles] (confidence: observed)
- Reduced motion is designed in: 62 `@media (prefers-reduced-motion:no-preference)` blocks gate the motion, and 17 `reduce` blocks. Navigation durations drop to `0s` under reduce — [source: CSS bundles] (confidence: observed)
- The homepage hero uses a script-rendered "wave" (`.single-wave`) with static `.webp` fallbacks per breakpoint, and rolling tabular figures in the hero eyebrow. The page also references Rive — [source: https://stripe.com/ HTML] (confidence: observed; animation internals not inspected)
- Stripe documents the following motion rules. Avoid built-in `ease-in`, `ease-out` and `linear` curves. Most animations should stay under 500ms. Animate only `transform` and `opacity`. Respect `prefers-reduced-motion`. Use IntersectionObserver rather than scroll listeners. One example used the Web Animations API with `duration: 800` and `easing: cubic-bezier(.2,1,.2,1)` — [source: https://stripe.com/blog/connect-front-end-experience] (confidence: documented, 2017)
- Elements exposes `disableAnimations: boolean` and a `transition` property in rules — [source: Appearance API docs; `@stripe/stripe-js` types] (confidence: documented)
- The Lenny and Friends Summit opening animation, designed by Dill's team, began from a 3D model. AI generated variations, which were reviewed frame by frame through 56 iterations — [source: Katie Dill, Lenny's Podcast / Lenny and Friends Summit transcript, YouTube GLvFTMtw4Jk] (confidence: documented)

## 5. Illustration and gradient style (described, not copied)

- The brand mark is a parallelogram, a slanted four-sided shape. In ads it is paired with users' own visuals — [source: Katie Dill transcript, YouTube GLvFTMtw4Jk] (confidence: documented). Not reproduced here.
- The wordmark is a lowercase, geometric sans word logo. On stripe.com its fill is themable through a `LogoColor` custom property (navy `#0A2540` by default on legacy pages) — [source: https://stripe.com/newsroom/information inline CSS/SVG] (confidence: observed). Not reproduced here.
- Visual language: light backgrounds (white, or `#f8fafd`) with blue-grey neutrals, against saturated gradient accents that run purple → magenta → orange/yellow. Gradients appear as text fills on numbers, thin gradient borders on inputs (a "shimmer" keyframe), blobs, and a large flowing "wave" in the hero — [source: homepage CSS and HTML] (confidence: observed for components; inferred for the overall description)
- Long-running motif: diagonal skewed stripe bands built with `transform: skewY(-12deg)` over CSS grid (2017) — [source: https://stripe.com/blog/connect-front-end-experience] (confidence: documented). Their current use was not verified — (inferred)
- Product imagery is CSS- or SVG-drawn UI (browser frames with dot controls via `--browser-graphic-dots-color`, cards, mini UI strokes `--hds-canary-ui-stroke`) rather than screenshots — [source: homepage CSS] (confidence: observed for tokens; inferred for the generality)
- Stripe Press uses a separate editorial identity. The *Built to Grow* cover was made by a human marbler, with AI then refining the colours and lines — [source: Katie Dill transcript, YouTube GLvFTMtw4Jk; https://press.stripe.com/] (confidence: documented)
- Dill's team described one ad critique that produced about 17 small fixes, including softer edges and smaller bubbles. Dill says the term "Pepsi bubbling" came out of that critique and is now used as a verb for meticulous craft — [source: Katie Dill transcript, YouTube GLvFTMtw4Jk] (confidence: documented; the origin of the name is not explained in the transcript)

## 6. Writing and voice

- No public, general Stripe writing or voice style guide was found — [source: searches of stripe.com and docs.stripe.com; docs index https://docs.stripe.com/llms.txt] (confidence: inferred from absence)
- Stripe Apps empty-state rules. Titles state what is missing, as a short phrase with a period ("No successful payments."), not a promotion. Descriptions explain when data appears, in under 14 words, in active voice. The action text echoes the title ("No customers yet." → "Add customer", not "Get started"). "No transactions yet" is preferred over "No transactions" — [source: https://docs.stripe.com/stripe-apps/patterns/empty-state.md] (confidence: documented)
- Onboarding should be "Effortless", "Customizable" and "Relevant". It should be brief, with one primary action per step. Credentials should be collected on the external site, never inside the Dashboard. Users should always be redirected back to Stripe — [source: https://docs.stripe.com/stripe-apps/patterns/onboarding-experience.md] (confidence: documented)
- Sentence case and verb-first headings run throughout the docs ("Customize appearance", "Create an Account Link", "Handle verification errors") — [source: docs pages cited above] (confidence: observed)
- Marketing voice: plain, declarative, benefit-first headline, e.g. "Financial infrastructure to grow your revenue." — [source: https://stripe.com/ h1] (confidence: observed)
- Optimism is a stated brand value that shapes writing — [source: Katie Dill transcript, YouTube GLvFTMtw4Jk] (confidence: documented)
- Docs are authored in Markdoc (open source, MIT), so conventions are encoded in tooling — [source: https://github.com/markdoc/markdoc] (confidence: observed for the repo; inferred for "conventions encoded")

## 7. Product UI patterns, templates and flows

### Account setup / Dashboard onboarding

- Flow: (1) create an account at `dashboard.stripe.com/register`; (2) work in a sandbox right away; (3) complete the account checklist; (4) verify the business at `/account/onboarding` (business info, product, your relationship to the business), which is required for live mode under KYC; (5) set public details (name, URL, support contact, statement descriptor); (6) secure the account with 2FA (passkeys or security keys recommended) and team invites — [source: https://docs.stripe.com/get-started/account/activate.md] (confidence: documented)
- Dashboard information architecture: the sidebar has a primary section (Home, Balances, Transactions, Customers, Product catalog), then Shortcuts (pinned and recent pages), then Products (Connect, Payments, Billing, Reporting, …), with settings split into Account and Product. Home has customisable widgets (Add → select widgets → Apply) and surfaces notifications such as disputes — [source: https://docs.stripe.com/dashboard/basics.md] (confidence: documented)
- Payments list (Transactions): filterable and exportable, with statuses. Clicking a row opens a detail page. Organizations add a cross-account filter — [source: same] (confidence: documented)

### Checkout (hosted)

- Lifecycle: (1) the app creates a Checkout Session; (2) the customer is redirected to the Session URL on a Stripe-hosted page; (3) the customer enters payment details and pays; (4) the customer returns to the app; (5) the `checkout.session.completed` webhook drives fulfilment — [source: https://docs.stripe.com/payments/checkout/how-checkout-works.md?payment-ui=stripe-hosted] (confidence: documented)
- Three UIs on the Checkout Sessions API. Full page offers "15 configurable settings via brand settings" and a full order summary. The embedded form offers "70 configurable settings via the Appearance API" and a limited summary. Elements offers "Full CSS customization via the Appearance API" and no summary — [source: https://docs.stripe.com/payments/checkout.md] (confidence: documented)
- Branding settings (Dashboard or the `branding_settings` API): logo or icon, display name, `background_color`, `button_color`, `font_family` (curated list: Be Vietnam Pro, Bitter, Chakra Petch, Hahmlet, Inconsolata, Inter, Lato, Lora, M PLUS 1 Code, Montserrat, Nunito, Noto Sans/Serif, Open Sans, …) and `border_style` (e.g. `rectangular`) — [source: https://docs.stripe.com/payments/checkout/customization/appearance.md?payment-ui=stripe-hosted] (confidence: documented)
- Built-in features: wallets and Link, responsive mobile layout, SCA, CAPTCHA, card validation, error messaging, tax, Adaptive Pricing, localisation — [source: how-checkout-works] (confidence: documented)

### Payment Element

- Layouts: `tabs` (payment methods in a horizontal row), `accordion` with radio buttons (`spacedAccordionItems: false`), and `accordion` without radios (`spacedAccordionItems: true`). Wallets show by default. When combined with the Express Checkout Element, wallets move there to avoid duplication — [source: https://docs.stripe.com/payments/payment-element.md] (confidence: documented)

### Connect hosted onboarding

- Flow: (1) the platform sets its brand name, colour and icon in Connect settings (required); (2) it creates the account and prefills known info; (3) it chooses what to collect (optionally public details); (4) it creates an Account Link and redirects; (5) Stripe collects the requirements and handles verification errors; (6) the user returns via the `return_url`, or the `refresh_url` if the link expired; (7) the platform tracks status and uses `account_update` links later — [source: https://docs.stripe.com/connect/hosted-onboarding.md] (confidence: documented)

### Stripe Apps UI

- An app lives in a Dashboard drawer opened from a dock. Its parts are the app indicator (colour bar and icon, the only place an arbitrary brand colour is allowed, via `brandColor` and `brandIcon`), a header (name, view, external link, top actions), content, the App Marketplace entry, and the active-app icon — [source: https://docs.stripe.com/stripe-apps/design] (confidence: documented)
- Views: `ContextView` (the default, side by side with Stripe content), `FocusView` (blocking backdrop, for longer start-to-finish tasks; "default to ContextView") and `SettingsView`. Surfaces are detail pages, list pages and Home — [source: same] (confidence: documented)
- "Custom styling of UI elements is intentionally limited … to maintain platform consistency … and to ensure a high accessibility bar." Colours are restricted per element — [source: same] (confidence: documented)
- Patterns published: full-page apps, chart layout, lists, filter controls, onboarding, sign in, settings sign in, demo content, additional context, redirect to Stripe, sign out, back link, action buttons, communicating state, empty state, loading, progress stepping, waiting screens — [source: https://docs.stripe.com/stripe-apps/patterns.md] (confidence: documented)
- Progress stepping: use a `FocusView` for multi-step tasks, put the stepping controls in the footer, and make the final action a primary Button — [source: https://docs.stripe.com/stripe-apps/patterns/progress-stepping.md] (confidence: documented)
- Layout is stack-based: `stack: 'x' | 'y' | 'z'`, `gap`, `alignX`, `alignY`, `distribute`, `wrap`, with fractional widths (`1/2` … `11/12`, `fill`) and content sizing (`min`, `max`, `fit`). The docs say this is "the same layout system Stripe uses internally" — [source: https://docs.stripe.com/stripe-apps/style] (confidence: documented)

## 8. Public open-source Stripe UI code

- `@stripe/react-stripe-js` 7.0.0 (MIT, github.com/stripe/react-stripe-js) exports `Elements`, `ElementsConsumer`, `useStripe`, `useElements`, `EmbeddedCheckout`, `EmbeddedCheckoutProvider`, and element components such as `PaymentElement`, `ExpressCheckoutElement`, `AddressElement`, `LinkAuthenticationElement`, `ContactDetailsElement`, `CardElement`, `CurrencySelectorElement`, `TaxIdElement`, `PaymentMethodMessagingElement`, `TermsElement`, plus Issuing display elements. A separate `checkout` entry point also exists — [source: npm tarball `dist/react-stripe.d.ts`] (confidence: observed)
- These are wrappers. The rendered UI runs inside Stripe-hosted iframes, so no Stripe component CSS ships in the packages — (confidence: inferred, from wrapper exports and Appearance-only styling)
- `@stripe/stripe-js` 10.1.0 types: `theme?: 'stripe' | 'night' | 'flat'`, `labels?: 'auto' | 'above' | 'floating'`, `inputs?: 'spaced' | 'condensed'`, `disableAnimations?: boolean` — [source: npm tarball `dist/stripe-js/elements-group.d.ts`] (confidence: observed)
- Appearance API commonly used variables: `fontFamily`, `fontSizeBase`, `spacingUnit`, `borderRadius`, `colorPrimary`, `colorBackground`, `colorText`, `colorDanger` — [source: https://docs.stripe.com/elements/appearance-api.md?api-integration=paymentintents] (confidence: documented)
- Less common variables:
  - buttons: `buttonBorderRadius`, `buttonBoxShadow`, `buttonExpressCheckoutBorderRadius`, `buttonColorBackground`, `buttonColorText`, `buttonFontSize`, `buttonFontWeight`, `buttonPadding`
  - focus: `focusBoxShadow`, `focusOutline`
  - labels: `labelColorText`, `labelFontSize`, `labelFontWeight`, `labelSpacing`
  - inputs: `inputBoxShadow`, `inputColorBorder`, `inputFocusBoxShadow`, `inputFocusColorBorder`
  - type: `fontSmooth`, `fontVariantLigatures`, `fontVariationSettings`, `fontWeightLight`, `fontWeightNormal`, `fontWeightMedium`, `fontWeightBold`, `fontLineHeight`, `fontSize3Xs` … `fontSize2Xl`
  - logos: `logoColor`, `tabLogoColor`, `tabLogoSelectedColor`, `blockLogoColor`
  - status colours: `colorSuccess`, `colorWarning`, and `accessibleColorOn{ColorPrimary,ColorBackground,ColorSuccess,ColorDanger,ColorWarning}`
  - text: `colorTextSecondary`, `colorTextPlaceholder`
  - icons: about 25 `icon*` and `tabIcon*` colours
  - accordion: `accordionItem*` (spacing, trigger colours, label type)
  - grid and spacing: `gridColumnSpacing`, `gridRowSpacing`, `pickerItemSpacing`, `tabSpacing`
  - [source: same] (confidence: documented)
- Constraint: `colorPrimary`, `colorBackground`, `colorText`, `colorSuccess`, `colorDanger` and `colorWarning` do not accept `rgba()` or `var()` — [source: same] (confidence: documented)
- Rules target public classes only, with no descendant selectors. The classes are `.Tab/.TabIcon/.TabLabel` (`--selected`), `.Label` (`--empty/--invalid/--focused/--floating/--resting`), `.Input` (`--empty/--invalid`, `:hover/:focus/:disabled/:autofill`, `::placeholder/::selection`), `.Error`, `.Block/.BlockDivider/.BlockAction`, `.CodeInput`, `.Checkbox*`, `.Dropdown/.DropdownItem`, `.Switch/.SwitchControl`, `.PickerItem/.PickerAction`, `.Menu*`, `.AccordionItem`, `.PaymentMethodMessaging`, `.RadioIcon*`. Properties are allowlisted (box model, border, colour, font, `transition`, SVG `fill`, `stroke`, `r`, …) — [source: same] (confidence: documented)
- Stripe's own example `.Tab` shadow: `0px 1px 1px rgba(0,0,0,.03), 0px 3px 6px rgba(18,42,66,.02)`, with the selected state adding `0 0 0 2px var(--colorPrimary)`. Example vars: `colorPrimary '#0570de'`, `colorText '#30313d'`, `colorDanger '#df1b41'`, `spacingUnit '2px'`, `borderRadius '4px'` — [source: same] (confidence: documented; these are examples, not documented defaults)
- Connect embedded components have their own appearance set, about 90 variables (button primary/secondary/danger, action, badge, form, overlay, and heading/body/label type tokens). Defaults include `colorPrimary` `#0074D4`, `colorText` `#444444`, `colorSecondaryText` `#717171`, `colorDanger` `#DF1B41`, `fontSizeBase` 16px and `overlays` `dialog | drawer` — [source: https://docs.stripe.com/connect/embedded-appearance-options.md] (confidence: documented)
- `@stripe/ui-extension-sdk` 9.4.0 (MIT) is the public UI kit for Stripe Apps. Its components: Accordion, AccordionItem, Badge, Banner, BarChart, Box, Button, ButtonGroup, Checkbox, Chip, ChipList, Code, ContextView, CurrencyField, DataTable, DateField, DateRangePicker, DetailPage, DetailPagePropertyList, DetailPageTable, Divider, FocusView, FormFieldGroup, FullPageView, Icon, Img, Inline, LineChart, Link, List, ListItem, Menu, MenuGroup, MenuItem, OnboardingView, OverviewPage, PageModule, PlatformConfigurationView, PropertyList, PropertyListItem, Radio, SearchField, Select, SettingsView, SignInView, Sparkline, Spinner, StripeFileUploader, Switch, Tab/TabList/TabPanel/TabPanels/Tabs, Table*, TaskList, TaskListItem, TextArea, TextField, Tooltip — [source: npm tarball `ui/index.d.ts`] (confidence: observed)
- SDK token unions: colour `brand | primary | secondary | disabled | info | success | attention | critical`; background `container | surface`; keyline `neutral | critical`; Button `type` `primary | secondary | destructive` — [source: same] (confidence: observed)
- Other public repos: `stripe/stripe-apps` (sample apps, MIT), `stripe/elements-examples` (MIT), `stripe/stripe-react-native` (MIT), `markdoc/markdoc` (MIT) — [source: GitHub API metadata] (confidence: observed)

## 9. Public talks and writing on Stripe design

- **Katie Dill, "How to scale intent, quality, and artistry with AI"**, Lenny and Friends Summit (posted on the Lenny's Podcast channel, dated 2026-09-25 by the transcript source; the brief says the summit was 2026-09-10) — [source: https://www.youtube.com/watch?v=GLvFTMtw4Jk; https://www.lennysnewsletter.com/p/all-of-the-lenny-and-friends-summit] (confidence: documented). Each claim below was checked against the unofficial transcript at https://hraness.com/reading/raise-the-ceiling-how-to-scale-intent-quality-and-artistry:
  - "at Stripe, we care a lot about optimism", which applies to colours, writing, subjects and products for entrepreneurs — (confidence: documented)
  - "we started with an MCP that understood our design documentation, but the results were not great … Three different people could put in the same prompt and get three different results." — (confidence: documented)
  - "Now we have created a CLI built on our design system". The system is "not just components and atomic parts, but … full templates and flows". It loads docs at the right moment and makes the AI "far more obedient" — (confidence: documented)
  - "Pepsi bubbling" is used "as a verb to mean meticulous craft", meaning going one level deeper than customers can see — (confidence: documented)
  - The *Built to Grow* cover "was created by a human marbler", with AI refining colours and lines — (confidence: documented)
  - The summit opening animation took "56 iterations" — (confidence: documented)
  - "the old system scaled consistency, but the new system needs to scale intent" — (confidence: documented)
  - Editors take "accountability for every decision … every pixel". "We have to unlearn that built means done, and that done means good." — (confidence: documented)
  - Neither the CLI nor the MCP is named, and no metrics are given — (confidence: observed in transcript)
- **Lenny's Podcast, "Building beautiful products with Stripe's Head of Design | Katie Dill"** (2023-10-15). Its chapters cover quality tied to growth, defining beauty, operationalising quality, "Stripe's PQR quality review", and "15 Essential Journeys: Stripe's method for holistic UX understanding" — [source: https://www.lennysnewsletter.com/p/building-beautiful-products-with] (confidence: documented, show notes only)
- **How I AI, Owen Williams (Stripe design manager) on Protodash** (2026-05-04). Protodash is an internal prototyping studio built from Cursor rules, MCPs and the design system's React components. It addresses "blurple slop" (AI output that looks almost, but not quite, like Stripe) and adds design review modes, variants, self-testing screenshots and annotate-for-AI — [source: https://www.lennysnewsletter.com/p/the-internal-ai-tool-thats-transforming] (confidence: documented). Secondary coverage says it uses an MCP server for **Sail** — [source: https://www.chatprd.ai/how-i-ai/stripe-owen-williams-on-buildling-internal-prototyping-studio] (confidence: inferred; secondary source)
- **"Designing accessible color systems"** (Daryl Koopersmith and Wilson Miner, 2019-10-15). Covers CIELAB-based ramps, the 5-level and 4-level contrast rules, and an internal palette editor that shows impossible colours — [source: https://stripe.com/blog/accessible-color-systems] (confidence: documented)
- **"Connect: behind the front-end experience"** (Benjamin De Cock, 2017-06-19). Covers CSS 3D cubes, skewed grid stripes, Web Animations API, and the motion performance rules in §4 — [source: https://stripe.com/blog/connect-front-end-experience] (confidence: documented)
- **Stripe Press**: a publishing imprint with its own editorial design. Its current list includes *Built to Grow: A Handbook for High-Performance Sales Teams* by Stephanie Friedman, *Scaling People*, *Poor Charlie's Almanack*, *The Scaling Era* and others — [source: https://press.stripe.com/] (confidence: observed)

## 10. Brand guidelines and the legal boundary

- No public brand guideline document (clear space, colour specs, typography rules for third parties) was found. The newsroom offers a "Download logo kit" ZIP and a "Powered by Stripe" badge ZIP, plus the note "All Stripe trademarks, logos, or other brand elements are subject to" the Mark Usage Terms — [source: https://stripe.com/newsroom/information] (confidence: observed; ZIPs deliberately not downloaded)
- "Marks" are defined broadly, including trade dress and design elements used to identify Stripe's goods or services, registered or not — [source: https://stripe.com/legal/marks] (confidence: documented)
- Permitted: use the Marks only where they directly relate to Stripe services (e.g. a checkout page), follow Stripe's style guidelines, and use the word mark unaltered to refer accurately to Stripe — [source: same] (confidence: documented)
- Forbidden:
  - altering the Marks (abbreviating, pluralising, translating)
  - implying endorsement or a partnership, or showing Stripe's Marks more prominently than your own
  - using them on merchandise
  - combining them with something into a new composite mark
  - incorporating them into your own trademark, trade dress, product or domain name, or social handle, or using confusingly similar ones
  - disparaging use
  - adding ™ or ®
  - [source: same] (confidence: documented)
- Contact for other uses: trademarks@stripe.com — [source: same] (confidence: documented)
- Stripe Apps may use any colour only in the app indicator. Everything else is constrained to Stripe's tokens — [source: https://docs.stripe.com/stripe-apps/design] (confidence: documented)
- Implication for this project. A reference built from these observations should:
  - use its own name, logo and parallelogram-free marks
  - avoid presenting itself as Stripe or as Stripe-endorsed
  - not bundle Söhne or any Stripe asset files
  - avoid reproducing Stripe's overall trade dress closely enough to cause confusion
  Token values may be documented as observations, but a shipped theme should be derivative and clearly labelled as "inspired by" — (confidence: inferred, from the Marks definition including trade dress; not legal advice)

## 11. What is not publicly available

- **Sail**, the internal design system and component library "that all of Stripe's products are built on". Its source, tokens, docs and Figma libraries are not public — [source: Stripe job listing as quoted in search results (https://stripe.com/jobs/listing/product-designer-design-systems/6865619, now 404); How I AI coverage] (confidence: inferred, based on secondary quotes and a dead listing)
- The design-system CLI with full templates and flows (Dill), and the earlier design-docs MCP. Neither is named, released or documented — [source: Katie Dill transcript] (confidence: documented as existing; not public)
- Protodash (internal prototyping studio) — [source: How I AI] (confidence: documented as internal)
- Internal process artefacts: the PQR quality review, the 15 Essential Journeys, and the 17-point critique checklists are referenced but not published — [source: Lenny's Podcast show notes; Dill transcript] (confidence: documented as existing)
- Dashboard UI tokens and computed styles need a logged-in session. They were not observed here, and anything about Dashboard visuals beyond the docs would be guesswork — (confidence: inferred)
- The Stripe Apps colour and type token values (hex, px) behind names like `brand`, `critical` and `heading` are absent from the public `.md` docs and the SDK types — (confidence: observed absence)
- The default variable values of the Elements `stripe`, `night` and `flat` themes are not documented — (confidence: observed absence)
- Figma: the `@stripedesign` Figma Community toolkit returned 403 here. Its contents and licence are unverified, and the current Apps design page no longer mentions it — [source: https://www.figma.com/@stripedesign; https://docs.stripe.com/stripe-apps/design] (confidence: observed)
- The internal palette editor (2019 post), the motion guidelines beyond the 2017 post, iconography rules, illustration and art-direction guides, and the hero wave's shader parameters are unpublished — (confidence: inferred from absence)
- Söhne licence: Stripe's font file is licensed to Stripe and is not reusable — (confidence: inferred)
