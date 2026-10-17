---
status: accepted
---

# Write the server in Go on the official MCP Go SDK, stdio transport first

The server is written in Go using the official `github.com/modelcontextprotocol/go-sdk`, run over stdio, and distributed as a single static binary (via `go install` and prebuilt release binaries) rather than an npm package. A single binary needs no Node runtime on the user's machine, which suits a tool that runs next to Claude Code or Claude Desktop. Remote (HTTP) transport is deferred until a hosted use case exists (see ADR-0005). This supersedes ADR-0001 (TypeScript).
