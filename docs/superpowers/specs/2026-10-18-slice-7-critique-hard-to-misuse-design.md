# Slice 7: Make `critique` hard to misuse — design

Approved: 2026-10-18

Self-approved under the project's no-approval-gates rule (see CLAUDE.md, "Decide, don't ask"). The human asked for this fix after a real session. Path: bounded (one tool handler, its input type and tests; no new module and no interface other tools depend on).

## Goal

A client that works in a different folder from the server's `--dir` can call `critique` correctly on its first try, and when it still gets a path wrong, the error tells it how to fix the call.

## Evidence

A session in `tenet-ui/sample` used a server started with `--dir …/packs/stripe-reference/`. The model called:

1. `critique(file: "payments-list.html", system: ".")` and got `file "payments-list.html" not found.` A relative `file` resolves against the server root (the pack folder), not the model's working folder.
2. `critique(file: <absolute path>, system: "<the sample folder>")` and got `design-system not found under …/sample; run tenet init there.` The model read "folder holding design-system/" as "my project", and `system` resolves against the server root too.

The model then refused to run `tenet init` (rightly) and checked the rules by hand with `rg`, losing the rule ids and fixes `critique` returns.

## Acceptance criteria (from docs/roadmap.md)

1. `critique` works with `file` alone: an empty or missing `system` means `.`.
2. A "design-system not found" or "file not found" error names the absolute path searched and the server root, and says how to fix the call. The design-system error no longer tells the model to run `tenet init`.

## Decisions

| # | Decision | Reason |
|---|---|---|
| M1 | `CritiqueInput.System` gets `json:"system,omitempty"`, so the tool's input schema no longer lists it as required. An empty or missing `system` is treated as `.`. The `system is empty` error is removed. | In a server started with `--dir`, the served system is nearly always the one meant. A required parameter that is nearly always `.` is the trap. The SDK validates input against the schema before the handler runs, so `omitempty` is needed for a missing field to reach the handler. |
| M2 | `system` stays accepted, relative to the server root or absolute, for a pack that sits in a subfolder of the root. | Slice 4 tests and the existing README use `system` to select a folder under the root. Removing it would break them for no gain. |
| M3 | `system`'s `jsonschema` description becomes: `optional; . or empty means the design system this server was started with. Do not pass your project folder.` `file`'s becomes: `path to the UI source file; relative paths resolve against the server root, so pass an absolute path for a file in your own working folder`. The tool description in `server.go` is rewritten to match and no longer says `. for the project root`. | The parameter text is the only guidance a model reads before the first call. |
| M4 | File not found: `file "<file>" not found. Relative paths resolve against the server root <abs root>; pass an absolute path for a file outside it.` | Names the root and the fix. The file argument is echoed as given, as today. |
| M5 | Design-system not found (the `directory not found` problem only): `design-system not found under <abs system root>. This server's root is <abs root>; omit system to check against the system it serves.` No `tenet init` text. | `tenet init` writes files into the user's project and a model cannot act on it mid-task. The human can still read the server's startup line. |
| M6 | Other design-system problems (for example `not a directory`) keep their current text and the `; run tenet init` suffix. The delivery tools are unchanged. | Out of scope. The evidence shows only the not-found case, and delivery tools have no `system` argument to misuse. |
| M7 | `<abs root>` is `filepath.Abs(root)`; if that fails, the root as given. The absolute system root is `filepath.Join` of the absolute root and `system`, or `system` itself when absolute, cleaned. | Absolute paths make the message unambiguous. A relative root (the tests use `../..`) would not. |
| M8 | The result header (`No findings in <file> against <system>`) shows `.` when `system` was omitted, as it does for an explicit `.`. | One code path: the default is applied before the rest of the handler runs. |
| M9 | README: the `critique` section states that `system` is optional, that paths resolve against the server root, and that the model should pass an absolute `file` for files elsewhere. | The README is the human's documentation of the tool. |

## Tests (internal/server/critique_test.go)

- Update `TestCritiqueNoDesignSystemIsError` to the M5 message (with `t.TempDir()` as the root, assert the absolute root appears and `tenet init` does not).
- Replace the `empty system` case in `TestCritiqueEmptyArgumentsAreErrors` (it now succeeds): a new test calls `critique` with `file` only, and with `system: ""`, and gets the same result as `system: "."`.
- New test: a missing relative file returns the M4 message with the absolute root.
- New test: a call whose `system` is an absolute folder with no `design-system/` returns M5 naming that folder and the server root.
- New test: the registered tool's input schema does not list `system` as required (read it through the client's `ListTools`).
- Existing finding, no-finding, binary, too-large and non-regular-file tests are unchanged.

## Out of scope

- A server-wide `Instructions` string that tells clients when to call tools (suggested separately; its own slice).
- Resolving relative `file` paths against the client's working folder. The server cannot know it.
- Changing the delivery tools' errors.
