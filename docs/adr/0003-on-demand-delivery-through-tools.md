# Deliver design-system context on demand through MCP tools, not one preloaded blob

Rather than injecting the whole system into every prompt, the server exposes tools that return only what the current task needs: the tokens for a component, the rules for a template, the flow a user is in. Loading everything at once causes context rot, which is the failure the talk warns about ("consumes the documentation at the right time and place"). The cost is that the server must index the files and keep tool responses small.

## Result (2026-10-18)

The consistency experiment (slice 6, [experiments/consistency/REPORT.md](../../experiments/consistency/REPORT.md)) ran five briefs in both conditions. Pre-registered answer: MIXED. Tools produced fewer critique findings (53 against 184) but no output was token-only under the pre-registered metric. A post-hoc split shows every tools finding is a token declaration in the output's `:root` block, and no tools output has a raw colour outside it. The decision is not superseded: on that post-hoc split, tool delivery is supported at the use sites (the plain outputs carry 68 raw use-site values, the tools outputs none).
