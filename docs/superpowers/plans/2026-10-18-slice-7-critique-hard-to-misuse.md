# Slice 7: Make `critique` hard to misuse — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Do not commit; leave changes in the working tree.

**Goal:** `critique` works with `file` alone, and its path errors name the absolute path searched and the server root.

**Architecture:** All behaviour lives in `critiqueTool` (`internal/server/critique.go`). `CritiqueInput.System` becomes optional and is defaulted to `.` at the top of the handler. Two error strings change and gain the absolute server root from one helper. No new module.

**Tech Stack:** Go, `github.com/modelcontextprotocol/go-sdk` v1.8.0.

**Spec:** `docs/superpowers/specs/2026-10-18-slice-7-critique-hard-to-misuse-design.md`

## Global Constraints

- Decisions M1 to M9 in the spec; message texts in M4 and M5 are exact (substitute the angle-bracket values).
- The delivery tools and their errors do not change (M6).
- `test -z "$(gofmt -l .)"` and `go test ./...` pass; golangci-lint v2.15.0 is clean.
- Stdout carries only MCP protocol while serving; no new output.

## Review Focus

- `system` omitted entirely (not just `""`): must reach the handler and behave as `.` (schema must not require it).
- Server root passed as a relative path (the tests use `../..`): messages must still show an absolute root.
- Absolute `system` that points at a folder without `design-system/`: message names that folder and the root, not `tenet init`.
- A real existing `design-system/` that fails with another problem (for example `not a directory`) keeps its old text and suffix (M6).
- Explicit `system: "."` still behaves exactly as before, including the result header.

---

### Task 1: Optional `system`, clearer errors, new descriptions

**Model:** sonnet

**Files:**
- Modify: `internal/server/critique.go`
- Modify: `internal/server/server.go` (the `critique` tool `Description`)
- Test: `internal/server/critique_test.go`

**Interfaces:**
- Consumes: `designsystem.Load(dir) (DesignSystem, []Problem)`, `problemFor`, `toolError`, `textResult` (existing in package `server`).
- Produces: `CritiqueInput{File string "json:\"file\""; System string "json:\"system,omitempty\""}`; unexported `absRoot(root string) string` returning `filepath.Abs(root)` or `root` on error.

- [ ] **Step 1: Write the failing tests** in `critique_test.go`:
  - `TestCritiqueOmittedSystemDefaultsToRoot`: root is `critiqueRoot` (existing fixture) with a file that has a known finding. Call with `{"file": F}` (no `system` key) and with `{"file": F, "system": ""}`; each result must equal the result of `{"file": F, "system": "."}` (same text, `IsError` false). Use the fixture and file the existing happy-path test uses.
  - `TestCritiqueSystemIsNotRequired`: `session.ListTools`; find `critique`; its `InputSchema` `required` list contains `file` and does not contain `system`.
  - `TestCritiqueFileNotFoundNamesRoot`: temp root; call `{"file": "missing.html"}`; text equals ``file "missing.html" not found. Relative paths resolve against the server root <abs root>; pass an absolute path for a file outside it.`` with `<abs root>` = `t.TempDir()` value (already absolute).
  - `TestCritiqueRelativeRootIsShownAbsolute`: `connect(t, "../..")`, `{"file": "missing.html"}`; the text contains `filepath.Abs("../..")` and does not contain `root ../..`.
  - `TestCritiqueAbsoluteSystemWithoutDesignSystem`: temp root `r` with `page.html`; second temp dir `other` (no `design-system/`); call `{"file": "page.html", "system": other}`; text equals ``design-system not found under <other>. This server's root is <r>; omit system to check against the system it serves.``
  - Rewrite `TestCritiqueNoDesignSystemIsError` to expect the same M5 text for `system: "."` with `<root>` for both paths (`Join(root, ".")` cleans to `root`), and assert the text does not contain `tenet init`.
  - In `TestCritiqueEmptyArgumentsAreErrors` delete the `empty system` case; keep `empty file`.

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/server -run Critique -v`
Expected: the new and rewritten tests FAIL (old messages, `system` still required).

- [ ] **Step 3: Implement in `critique.go`**
  - `CritiqueInput.System` tag becomes `json:"system,omitempty"` with `jsonschema:"optional; . or empty means the design system this server was started with. Do not pass your project folder."`; `File`'s `jsonschema` becomes `path to the UI source file; relative paths resolve against the server root, so pass an absolute path for a file in your own working folder`.
  - Remove the `system is empty` check; instead `if in.System == "" { in.System = "." }` before the file checks.
  - Add `absRoot(root string) string`.
  - File-not-found message per spec M4, using `absRoot(root)`.
  - In the `directory not found` branch, return the M5 message: the system path is `in.System` when absolute else `filepath.Join(absRoot(root), in.System)` (`filepath.Clean` applied). The other-problem branch keeps its current code (M6).
  - In `server.go` rewrite the `critique` tool `Description` so it says `system` is optional and defaults to the system the server serves, and that relative paths resolve against the server root.

- [ ] **Step 4: Run to verify they pass**

Run: `go test ./... && test -z "$(gofmt -l .)" && go vet ./...`
Expected: all PASS, no gofmt output.

---

### Task 2: README

**Model:** haiku

**Files:**
- Modify: `README.md` (the `critique` section around lines 122 to 135, and any sentence there that says `system` is required)

**Interfaces:** none.

- [ ] **Step 1:** In the `critique` section, state in plain words: `file` is required and `system` is optional (omit it to check against the system the server was started with); relative paths resolve against the server's `--dir`, so pass an absolute `file` for a file in another folder; do not pass a project folder as `system`. Keep the existing example output and the other paragraphs unchanged.
- [ ] **Step 2:** Run `rg -n "system" README.md` and confirm no remaining sentence calls `system` required or says `. for the project root`.
- [ ] **Step 3:** Run `go test ./...`; Expected: PASS (docs only, nothing breaks).
