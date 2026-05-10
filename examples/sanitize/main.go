// Package main demonstrates calling [mcpsanitize.Sanitize] directly on a
// synthetic polluted args map.
//
// Run with:
//
//	go run ./examples/sanitize
//
// The example builds a tool-call argument map that exhibits all four
// pollution patterns the library detects, hands it to Sanitize, and prints
// both the cleaned args and the resulting Report.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	mcpsanitize "github.com/hollis-labs/go-mcp-sanitize"
)

func main() {
	// A polluted memory_write args map: payload_summary contains a
	// trailing self-named close-tag plus a leaked sibling block; the
	// sibling slot (payload_body) is empty so recovery is allowed.
	args := map[string]any{
		"namespace":  "user/example/memory",
		"memory_key": "decisions.example.demo",
		"payload_summary": "Sample summary text that runs to a natural " +
			"sentence boundary.</payload_summary>\n" +
			`<parameter name="payload_body">## Decision

Recovered body content from a leaked sibling block.</parameter>`,
		"payload_body": "",
		"tags":         `["decision", "example"]`,
	}

	cleaned, report := mcpsanitize.Sanitize(args)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	fmt.Println("# cleaned args")
	if err := enc.Encode(cleaned); err != nil {
		fmt.Fprintf(os.Stderr, "encode cleaned: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("# report")
	if err := enc.Encode(struct {
		Changed          bool              `json:"changed"`
		FieldsCleaned    []string          `json:"fields_cleaned"`
		RecoveredFields  map[string]string `json:"recovered_fields"`
		DroppedFragments []string          `json:"dropped_fragments"`
	}{
		Changed:          report.Changed(),
		FieldsCleaned:    report.FieldsCleaned,
		RecoveredFields:  report.RecoveredFields,
		DroppedFragments: report.DroppedFragments,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "encode report: %v\n", err)
		os.Exit(1)
	}
}
