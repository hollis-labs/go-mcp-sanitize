# go-mcp-sanitize

[![Go Reference](https://pkg.go.dev/badge/github.com/hollis-labs/go-mcp-sanitize.svg)](https://pkg.go.dev/github.com/hollis-labs/go-mcp-sanitize)

A small Go module for cleaning malformed Anthropic tool-call XML that leaks
into the **values** of MCP free-text parameters, plus a one-line middleware
adapter for [`github.com/mark3labs/mcp-go`][mcp-go].

## The bug it fixes

Some agent harnesses (Claude Code among others) occasionally emit a stray
`</PARAM_NAME>` close-tag — or a full `<parameter name="OTHER">…</parameter>`
block — *inside* the value of a free-text MCP tool-call argument. The MCP
server parses the call correctly, but stores the polluted JSON.

Concretely, the `payload_summary` field of a memory-write tool call ends up
looking like:

```
…concluding sentence.</payload_summary>
<parameter name="payload_body">## Decision

…full body content…</parameter>
```

while the *separate* `payload_body` field is already correctly populated by
the same call. Each polluted write then needs a second supersede call to
overwrite the bad copy — observable as a ~2× write-amplification on any
free-text tool over a long session.

This library detects four common pollution shapes and produces a cleaned
`args` map. The library itself is zero-side-effect (no logs, no globals).
Telemetry is the responsibility of the optional middleware.

## Status

Pre-1.0 / experimental. The public API surface is small and stable, but
v0.x bumps may still carry breaking changes — they will be called out
loudly in [`CHANGELOG.md`](./CHANGELOG.md).

## Install

```bash
go get github.com/hollis-labs/go-mcp-sanitize@latest
```

Floor versions: Go 1.26.2, `github.com/mark3labs/mcp-go` v0.47.0.

Runnable demos live under [`examples/`](./examples).

## Use the middleware (5 lines)

```go
import (
    "log/slog"

    "github.com/hollis-labs/go-mcp-sanitize"
    "github.com/mark3labs/mcp-go/server"
)

logger := slog.Default()
srv := server.NewMCPServer("my-server", "1.0.0")
srv.AddTool(tool, mcpsanitize.Middleware(logger)(handler))
```

The middleware extracts the args, runs `Sanitize`, and replaces
`request.Params.Arguments` with the cleaned map before invoking your handler.
Clean calls pass through unmodified and emit no log line. Polluted calls
emit exactly one warn-level line:

```
mcp-sanitize: cleaned tool call  tool=memory_write  fields_cleaned=[payload_summary]  recovered_fields=[payload_body]  dropped_count=2
```

To detect the report rate in production, count those warn lines per server
per day — that's your "cleaned-call rate" metric.

## Detect manually

If you want to clean args without the middleware:

```go
cleaned, report := mcpsanitize.Sanitize(args)
if report.Changed() {
    // … log, count, whatever
}
// hand `cleaned` to your handler
```

The `Report` struct exposes the three signals you need for telemetry:
`FieldsCleaned`, `RecoveredFields`, `DroppedFragments`.

## The four detection patterns

Applied per field, in order:

1. **Trailing self-named close-tag** — value ends with `</PARAM_NAME>` (or
   has it followed only by whitespace / short markup). Strip from the marker
   onward.
2. **Leaked sibling `<parameter>` block** — value contains
   `<parameter name="OTHER">CONTENT</parameter>` (or the open-only tail
   variant). Extract `CONTENT`. If `args["OTHER"]` is empty/missing, inject
   it (and record in `Report.RecoveredFields`); if `args["OTHER"]` already
   has a non-empty value, drop the markup but **do not overwrite**. Either
   way the markup is removed from the source field.
3. **Trailing generic close-tag** — value ends with `</something>` near the
   tail that doesn't match the field name. Strip if it looks like leakage
   (no balancing open-tag within 64 characters earlier).
4. **Tags-array recovery** — if a `tags` field arrives as a string, try
   `json.Unmarshal` first; if that fails, run patterns 1–3 against it and
   try again; otherwise drop it (set `tags = []`). If `tags` arrives as a
   string array already, leave it alone.

## Public API

```go
package mcpsanitize

func Sanitize(args map[string]any) (cleaned map[string]any, report Report)
func CleanFreeText(value, paramName string) (clean string, recovered map[string]string, changed bool)
func Middleware(logger *slog.Logger) func(next server.ToolHandlerFunc) server.ToolHandlerFunc

type Report struct {
    FieldsCleaned    []string
    RecoveredFields  map[string]string
    DroppedFragments []string
}
func (r Report) Changed() bool
```

The API surface is locked — downstream MCP-server consumers can import it
as-is and follow standard semver guarantees for any future v0.x bump.

## Out of scope

- Backfilling existing polluted memories (separate ticket).
- Tightening MCP tool descriptions to discourage the pollution at source.
- "Request validation" beyond the four patterns above.

## License

MIT — see `LICENSE`.

[mcp-go]: https://github.com/mark3labs/mcp-go
