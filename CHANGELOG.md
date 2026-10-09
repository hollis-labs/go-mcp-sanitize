# Changelog

## Repository retirement — 2026-10-09

### Changed

- Maintained development moved to [github.com/hollis-labs/libs/plugin-mcp/go-mcp-sanitize](https://github.com/hollis-labs/libs/tree/plugin-mcp%2Fv0.1.1/plugin-mcp/go-mcp-sanitize) in
  `github.com/hollis-labs/libs/plugin-mcp@v0.1.1` (`plugin-mcp/v0.1.1`).
- This standalone repository is retired after the replacement release was
  verified fetchable with successful module CI. README migration instructions
  identify the new import prefix; existing standalone tags and history are preserved.

All notable changes to `github.com/hollis-labs/go-mcp-sanitize` are documented
in this file. The project follows [Semantic Versioning](https://semver.org/).

## v0.1.1 — 2026-05-10

Release-prep / public-OSS hygiene pass. No public API changes.

### Changed

- `go.mod` `go` directive bumped from **1.24 → 1.26.2** to track the current
  toolchain floor and to clear the four stdlib `crypto/x509` + `crypto/tls`
  CVEs reported against 1.26.1 (all transitively reachable via the
  `mark3labs/mcp-go` server import in the new examples). Consumers on Go
  1.24 / 1.25 should pin to v0.1.0.
- README rewritten to drop internal portfolio references; added a Status
  banner and a pointer to the new `examples/` directory.
- CHANGELOG ticket-ID footnote removed from the v0.1.0 entry.

### Added

- `doc.go` — dedicated package doc for godoc / pkg.go.dev rendering.
- `examples/middleware/` — runnable demo wiring `Middleware` into an
  `mcp-go` server.
- `examples/sanitize/` — runnable demo of the `Sanitize` function on a
  synthetic polluted args map.

### Internal

- `testdata/polluted-decision-memory.json` replaced with a synthetic
  equivalent that preserves the same pollution shape without leaking any
  real-world content.

## v0.1.0 — 2026-05-09

Initial release.

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
