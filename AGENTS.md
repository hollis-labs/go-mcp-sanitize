# go-mcp-sanitize

Cleans malformed Anthropic tool-call XML that leaks into the *values* of
free-text MCP parameters — a stray `</param_name>` close tag or a whole
`<parameter name="other">…</parameter>` block ending up inside a string field.
It detects four pollution shapes and returns a cleaned `args` map. The library
itself has no side effects: no logs, no globals, no mutation of the input.
Telemetry belongs to the optional `mcp-go` middleware.

## Start Here

- `README.md` describes the bug, with the concrete `payload_summary` example
  that motivated the module.
- `sanitize.go` owns the four patterns and the change report.
- `middleware.go` is the `mark3labs/mcp-go` adapter and the only place that
  logs.
- `examples/sanitize` and `examples/middleware` are runnable.
- `testdata/` holds the polluted fixtures.

## Commands

```bash
gofmt -l .
go vet ./...
go test -race -count=1 ./...
```

There is no CI workflow or Makefile in this repo.

## Boundaries

Sanitizing must never invent or destroy content. Two rules carry that, and both
have named tests: a recovered sibling parameter is not written over a copy the
call already populated correctly
(`TestPattern2_DoesNotOverwriteExistingCleanCopy`), and legitimate inline tags
in prose are left alone (`TestPattern3_LeavesLegitimateInlineTags`). A greedier
matcher would corrupt honest content, which is worse than the bug being fixed.

The input map is never mutated — callers get a cleaned copy
(`TestSanitize_OriginalNotMutated`). Middleware depends on that to compare
before and after.

Non-string fields pass through untouched (`TestSanitize_NonStringField`). Only
free-text values are in scope.

The core package stays free of logging and global state so it can be called
from a hot path; if you need observability, wrap it in the middleware rather
than instrumenting `sanitize.go`.
