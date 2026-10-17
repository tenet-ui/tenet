# tenet

An MCP server that lets Claude users encode their own design system (point of view, tokens, components, templates, flows, rules) and delivers it to the AI at the right moment, so AI-built interfaces carry intent instead of generic "zombie UI". The Stripe design system, reverse-engineered from public sources, is the reference example.

The tool is at the walking-skeleton stage and has one tool, `get_point_of_view`.

## Requirements

Go 1.27.0 (the toolchain version in `go.mod`).

## Run

```bash
go run ./cmd/tenet --dir testdata/project
go build -o tenet ./cmd/tenet
```

Without `--dir`, the working directory is the project root.

## Check the stdio protocol

```bash
( printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"check","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_point_of_view","arguments":{}}}'; \
  sleep 3 ) | go run ./cmd/tenet --dir testdata/project
```

Stdin must stay open until the responses arrive, which is why the command sleeps; stdout carries only JSON-RPC responses, and the startup line goes to stderr.

## Connect Claude Code

```bash
claude mcp add --transport stdio tenet -- /absolute/path/to/tenet --dir /absolute/path/to/your/project
```

Then run `claude mcp list` and confirm `tenet` is connected.

The project needs `design-system/point-of-view.md`; write it first, before using the tool. see [testdata/project/design-system/point-of-view.md](testdata/project/design-system/point-of-view.md) for an example.

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
