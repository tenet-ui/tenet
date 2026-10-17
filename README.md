# tenet

An MCP server that lets Claude users encode their own design system (point of view, tokens, components, templates, flows, rules) and delivers it to the AI at the right moment, so AI-built interfaces carry intent instead of generic "zombie UI". The Stripe design system, reverse-engineered from public sources, is the reference example.

The tool has three commands: serving MCP over stdio (the default), `validate`, and `init`. It has five MCP tools: `get_point_of_view`, and four delivery tools that return one item at a time (see [Delivery tools](#delivery-tools)).

## Requirements

Go 1.27.0 (the toolchain version in `go.mod`).

## Run

```bash
go run ./cmd/tenet --dir testdata/project
go build -o tenet ./cmd/tenet
```

Without `--dir`, the working directory is the project root.

## Start a design system

```bash
go run ./cmd/tenet init --dir /path/to/your/project
```

This writes an empty, valid `design-system/` folder. It refuses to run if `design-system` already exists, and it writes nothing in that case.

## Check a design system

```bash
go run ./cmd/tenet validate --dir testdata/project
```

Each problem prints as `path:line: message`, so most terminals and editors make it a link. The command exits 0 when the system is valid, 1 when it has problems, and 2 for a usage or `--dir` error.

## The design-system folder

```
design-system/
├── point-of-view.md     required: the brand point of view, free Markdown, not empty
├── tokens.json          required: design values, see below
├── rules.md             required: checkable rules, see below
├── components/          optional: one Markdown file per component
├── templates/           optional: one Markdown file per template
└── flows/               optional: one Markdown file per flow
```

**`tokens.json`** is an object of categories. Each category is an object of token names. A value is a string or a number:

```json
{
  "color": {
    "ink": "#1d2433",
    "accent": "#2f6fed"
  },
  "space": {
    "unit": 4
  }
}
```

Category and token names are lowercase letters, digits and dashes, starting with a letter (`ink`, `touch-target`). A token's full name is `category.name`, for example `color.ink`.

**`rules.md`** holds one rule per `##` heading. The heading is the rule's id, lowercase letters and digits joined by dots or dashes. Each rule has a statement and exactly one `Reason:` line:

```markdown
## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
```

A rules file with no rules is valid.

A rule can name the items and token categories it applies to with one `Applies to:` line. Each entry is `kind:id`, where kind is `component`, `template`, `flow` or `tokens`, and for `tokens` the id is a category name:

```markdown
## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
Applies to: tokens:color, component:button
```

A rule without an `Applies to:` line is valid, but no delivery tool returns it. Every reference must name an item or category that exists; `validate` reports the ones that do not.

A rule can also carry a check. The `Forbid:` line is a Go regular expression (RE2 syntax); each match on a line of a UI source file is a finding. The `Fix:` line says how to fix it. A rule has both lines or neither:

```markdown
## color.tokens-only
Components take colours from tokens, never from raw hex values.
Reason: A raw value drifts from the palette the moment one screen is edited alone.
Forbid: #[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?)\(
Fix: Replace the raw colour with a colour token from tokens.json.
```

A pattern that does not compile, a repeated `Forbid:` or `Fix:`, or an empty one is a problem that `validate` reports. The rule is then not checked.

**`components/`, `templates/`, `flows/`** each hold Markdown files. The file name, without `.md`, is the item's id and must be lowercase kebab-case (`primary-button.md`). The first line must be a level-1 heading with the item's title:

```markdown
# Button

The primary action on a surface. Use one per view.
```

Files in those folders that do not end in `.md` are ignored.

See [testdata/project/design-system](testdata/project/design-system) for a complete example that passes `validate`.

## Delivery tools

| Tool | Argument | Returns |
|---|---|---|
| `get_tokens` | `category`, for example `color` | one `category.name: value` line per token, then the rules linked to the category |
| `get_component` | `id`, the file name without `.md` | the file's text, then the rules linked to the component |
| `get_template` | `id` | the same, for `templates/` |
| `get_flow` | `id` | the same, for `flows/` |

A rule appears under `## Linked rules` only when it is valid and names the item or category. An unknown id returns an error that lists the ids that exist. An item whose file has a problem returns that problem. Each call reads the design system again, so edits show without restarting the server. Keep each result small: a single item and its rules should stay well under 4 KB.

## Critique tool

`critique` checks one UI source file against the rules of a design system that have a `Forbid:` line. It returns one finding per match:

| Argument | Meaning |
|---|---|
| `file` | the source file, relative to the project root or absolute |
| `system` | the folder holding its `design-system/`, relative to the project root; `.` for the project root |

A finding reads:

```
testdata/critique/page.html:4: color.tokens-only: found "#ff0000". Replace the raw colour with a colour token from tokens.json.
```

The result names the file and the system, then one line per finding. When no rule can be checked, the result says so. The check is deterministic and reads the file as text, one line at a time, so a pattern cannot span lines and cannot judge layout or intent. The client that calls `critique` does that judgement, using the rules from the delivery tools. A `Forbid:` pattern is an approximation of its rule: the bundled spacing pattern flags every raw `px` value in `margin`, `padding` or `gap`, including values that are multiples of the unit. The source must be 1 MiB or less and a text file.

## Check the stdio protocol

```bash
( printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"check","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_point_of_view","arguments":{}}}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_component","arguments":{"id":"button"}}}'; \
  sleep 3 ) | go run ./cmd/tenet --dir testdata/project
```

Stdin must stay open until the responses arrive, which is why the command sleeps. The check returns the point of view (id 2) and one delivery tool result, the `button` component (id 3). Stdout carries only JSON-RPC responses. The startup line, which includes the design-system summary and any problems, goes to stderr.

## Connect Claude Code

```bash
claude mcp add --transport stdio tenet -- /absolute/path/to/tenet --dir /absolute/path/to/your/project
```

Then run `claude mcp list` and confirm `tenet` is connected. Run `tenet validate --dir /absolute/path/to/your/project` first, so the point of view and the other files are known to be valid.

## Develop

```bash
go test ./...
test -z "$(gofmt -l .)"
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
```

## Docs

- [GLOSSARY.md](GLOSSARY.md)
- [docs/roadmap.md](docs/roadmap.md)
- [docs/adr/](docs/adr/)

## Reference pack

A worked, unofficial Stripe-inspired design system lives in `packs/stripe-reference/`. See its [README](packs/stripe-reference/README.md).
