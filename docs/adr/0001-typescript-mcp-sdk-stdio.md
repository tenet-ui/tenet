---
status: superseded by ADR-0007
---

# TypeScript on the official MCP SDK, stdio transport first

The server is written in TypeScript on Node 22+ using the official `@modelcontextprotocol/sdk`, published as an npm package and run over stdio. Claude Code and Claude Desktop both launch stdio MCP servers directly, and the TypeScript SDK is the reference implementation with the widest client support. Remote (HTTP) transport is deferred until a hosted use case exists (see ADR-0005).
