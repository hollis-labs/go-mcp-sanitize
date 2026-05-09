package mcpsanitize

import (
	"context"
	"log/slog"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Middleware wraps a mark3labs/mcp-go ToolHandlerFunc with auto-sanitization
// and warn-level slog telemetry on any clean. Apply once at server setup:
//
//	srv := server.NewMCPServer(...)
//	srv.AddTool(tool, mcpsanitize.Middleware(logger)(handler))
//
// or wrap the registration call with a helper if your codebase has many tools.
//
// Behavior:
//   - The middleware extracts the args map via request.GetArguments(),
//     runs Sanitize, and rebuilds request.Params.Arguments with the cleaned
//     map before invoking next.
//   - If Sanitize reports no changes, the request is forwarded untouched and
//     no log line is emitted.
//   - If Sanitize reports changes, exactly one warn-level log line is emitted
//     before the handler runs.
//   - If logger is nil, the middleware uses slog.Default().
func Middleware(logger *slog.Logger) func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()
			if len(args) > 0 {
				cleaned, report := Sanitize(args)
				if report.Changed() {
					logger.Warn("mcp-sanitize: cleaned tool call",
						"tool", request.Params.Name,
						"fields_cleaned", report.FieldsCleaned,
						"recovered_fields", sortedKeys(report.RecoveredFields),
						"dropped_count", len(report.DroppedFragments),
					)
					request.Params.Arguments = cleaned
				}
			}
			return next(ctx, request)
		}
	}
}

// sortedKeys returns the map keys in sorted order. Stable output keeps log
// lines deterministic across runs.
func sortedKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
