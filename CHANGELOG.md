# Changelog

All notable changes to `github.com/hollis-labs/go-mcp-sanitize` are documented
in this file. The project follows [Semantic Versioning](https://semver.org/).

## v0.1.0 — 2026-05-09

Initial release. Phase 1 of the MCP Tool-Call Sanitization Rollout
(SP-20260509-0001 / CW-20260509-0031).

### Added

- `Sanitize(args map[string]any) (cleaned map[string]any, report Report)` —
  clones the args map, applies four detection patterns per field, and
  returns a structured `Report` describing what was changed.
- `CleanFreeText(value, paramName string) (clean string, recovered map[string]string, changed bool)` —
  per-field workhorse for callers that don't want the whole-args wrapper.
- `Middleware(logger *slog.Logger) func(next server.ToolHandlerFunc) server.ToolHandlerFunc` —
  one-line integration with `github.com/mark3labs/mcp-go` v0.47.0+.
- `Report` struct with `FieldsCleaned`, `RecoveredFields`,
  `DroppedFragments`, and a `Changed()` helper.

### Detection patterns

1. **Trailing self-named close-tag** — strip a value-tail `</PARAM_NAME>` and
   any short trailing markup.
2. **Leaked sibling `<parameter>` block** — extract `<parameter name="OTHER">CONTENT</parameter>`
   from the source value; inject `CONTENT` into `args["OTHER"]` only when
   that slot is empty/missing (never overwrite an existing clean copy).
3. **Trailing generic close-tag** — strip a stray `</xxx>` near the tail
   when there's no balancing open-tag within 64 characters earlier.
4. **Tags-array recovery** — promote a stringified JSON array (`'["a","b"]'`)
   back into `[]any{"a","b"}`; if cleanup + retry fails, drop to `[]`.

### Telemetry

`Sanitize` is zero-side-effect (no logs, no I/O). The `Middleware` adapter
emits exactly one warn-level `slog` line per cleaned call (clean calls are
silent). Log fields: `tool`, `fields_cleaned`, `recovered_fields`,
`dropped_count`.
