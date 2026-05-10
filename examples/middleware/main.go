// Package main demonstrates wiring [mcpsanitize.Middleware] into a
// `github.com/mark3labs/mcp-go` server.
//
// Run with:
//
//	go run ./examples/middleware
//
// The example registers a single trivial echo tool, wraps its handler with
// the sanitize middleware, and dispatches one polluted in-process tool call
// to show the warn-level slog line the middleware emits and the cleaned args
// the handler receives. There is no network listener — the handler is
// invoked directly via the wrapped function.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	mcpsanitize "github.com/hollis-labs/go-mcp-sanitize"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// A trivial handler that echoes its (cleaned) args back as JSON.
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		out, err := json.MarshalIndent(args, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal args: %w", err)
		}
		return mcp.NewToolResultText(string(out)), nil
	}

	// Construct an mcp-go server (not started — we invoke the wrapped
	// handler directly below). This step shows where the middleware
	// would fit into a real registration call:
	//
	//	srv.AddTool(tool, mcpsanitize.Middleware(logger)(handler))
	_ = server.NewMCPServer("example", "1.0.0")

	wrapped := mcpsanitize.Middleware(logger)(handler)

	// Polluted call: payload_summary has a trailing self-named close-tag
	// plus a leaked sibling block; payload_body is empty so the recovered
	// content gets injected.
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "memory_write",
			Arguments: map[string]any{
				"payload_summary": "Decision summary that ends with a " +
					"natural sentence.</payload_summary>\n" +
					`<parameter name="payload_body">## Decision

Body content recovered from the leaked sibling block.</parameter>`,
				"payload_body": "",
			},
		},
	}

	res, err := wrapped(context.Background(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "handler error: %v\n", err)
		os.Exit(1)
	}

	// Print whatever the handler emitted (the cleaned args).
	if len(res.Content) == 0 {
		fmt.Println("(no content)")
		return
	}
	if tc, ok := res.Content[0].(mcp.TextContent); ok {
		fmt.Println(tc.Text)
	}
}
