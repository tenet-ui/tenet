# Deliver design-system context on demand through MCP tools, not one preloaded blob

Rather than injecting the whole system into every prompt, the server exposes tools that return only what the current task needs: the tokens for a component, the rules for a template, the flow a user is in. Loading everything at once causes context rot, which is the failure the talk warns about ("consumes the documentation at the right time and place"). The cost is that the server must index the files and keep tool responses small.
