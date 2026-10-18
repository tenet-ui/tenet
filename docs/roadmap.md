# Roadmap

- [x] Slice 1: Walking skeleton
  - `go run ./cmd/tenet` launches a stdio MCP server; a client (Claude Code) lists one tool, `get_point_of_view`, that returns the text of `design-system/point-of-view.md` from the working directory
  - Repo has a Go `.gitignore`, one passing `go test`, `gofmt` and `golangci-lint` checks, CI running them on push, and a README with run instructions for Claude Code
- [x] Slice 2: Design system pack format and validation
  - A `design-system/` folder with tokens (JSON), rules (Markdown, one rule per entry with a reason), components, templates and flows is read by the server
  - A `tenet validate` command reports missing or malformed files with file and line, and exits non-zero on error
  - A `tenet init` command scaffolds an empty, valid `design-system/` folder
- [x] Slice 3: On-demand delivery tools
  - Tools `get_tokens`, `get_component`, `get_template`, `get_flow` return only the requested item and its linked rules, so a single call stays small (under 4 KB for the bundled example)
  - Each tool returns an error that names the missing item and lists the available ones
- [x] Slice 4: Critique tool
  - Tool `critique` takes a file path to a UI source file and the relevant system, and returns a list of findings, each naming the rule it breaks and a specific fix
  - Findings reference rule IDs from the rules file; a test fixture with two seeded violations reports both and no others
- [x] Slice 5: Reference pack: unofficial Stripe-inspired system
  - A `packs/stripe-reference/` folder built from `docs/research/stripe-design-system.md`, with tokens, rules, components, one template and one flow, all citing their research source
  - The pack ships no Stripe logos or wordmarks, and its README states it is an unofficial reverse engineering
  - Running `critique` against a fixture page that uses only pack tokens reports no findings
- [x] Slice 6: Consistency experiment
  - Five fixed briefs are run twice each: once with a plain prompt, once with the MCP tools enabled and the reference pack loaded
  - A committed report lists how many generated outputs used only pack tokens and how many critique findings each produced, and states whether tool delivery improved consistency (this settles the risk in ADR-0003)

## Later
- Hosted or shared mode with accounts and auth (needs its own ADR; see ADR-0005)
- Remote HTTP transport (see ADR-0007)
- Importing tokens from Figma or design-token JSON exports
- Publishing and discovering community packs
- Multi-context packs with a `GLOSSARY-MAP.md`-style index for large systems
- Cap the number of findings `critique` returns, with a "showing N of M" header (a 1 MiB file of matches can produce about 31 MB of text; the format needs a decision)
- Reconcile ADR-0002's "critique checklist" with the loader: the ADR lists one, but no slice reads or requires it (found in slice 5 review)
- Consistency check: exempt token-declaration blocks (`:root` custom properties that hold pack values) from the `color.tokens-only` token-only test; the slice 6 experiment marked every tools output not token-only because its token declarations match the raw-value pattern (see experiments/consistency/REPORT.md)
